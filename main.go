// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

// This is the main package for the `packer` application.

//go:generate go run ./scripts/generate-plugins.go
package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/go-uuid"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/pathing"
	pluginsdk "github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/hashicorp/packer-plugin-sdk/tmp"
	"github.com/hashicorp/packer/command"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/version"
	"github.com/mitchellh/cli"
	"github.com/mitchellh/panicwrap"
	"github.com/mitchellh/prefixedio"
)

func main() {
	// Call realMain instead of doing the work here so we can use
	// `defer` statements within the function and have them work properly.
	// (defers aren't called with os.Exit)
	os.Exit(realMain())
}

// realMain is executed from main and returns the exit status to exit with.
func realMain() int {
	var wrapConfig panicwrap.WrapConfig
	// When following env variable is set, packer
	// wont panic wrap itself as it's already wrapped.
	// i.e.: when terraform runs it.
	wrapConfig.CookieKey = "PACKER_WRAP_COOKIE"
	wrapConfig.CookieValue = "49C22B1A-3A93-4C98-97FA-E07D18C787B5"

	if inPlugin() || panicwrap.Wrapped(&wrapConfig) {
		// Call the real main
		return wrappedMain()
	}

	// Generate a UUID for this packer run and pass it to the environment.
	// GenerateUUID always returns a nil error (based on rand.Read) so we'll
	// just ignore it.
	UUID, _ := uuid.GenerateUUID()
	os.Setenv("PACKER_RUN_UUID", UUID)

	// Determine where logs should go in general (requested by the user)
	logWriter, err := logOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't setup log output: %s", err)
		return 1
	}
	if logWriter == nil {
		logWriter = io.Discard
	}

	packersdk.LogSecretFilter.SetOutput(logWriter)

	// Disable logging here
	log.SetOutput(io.Discard)

	// We always send logs to a temporary file that we use in case
	// there is a panic. Otherwise, we delete it.
	logTempFile, err := tmp.File("packer-log")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't setup logging tempfile: %s", err)
		return 1
	}
	defer os.Remove(logTempFile.Name())
	defer logTempFile.Close()

	// Setup the prefixed readers that send data properly to
	// stdout/stderr.
	doneCh := make(chan struct{})
	outR, outW := io.Pipe()
	go copyOutput(outR, doneCh)

	// Enable checkpoint for panic reporting
	if config, _ := loadConfig(); config != nil && !config.DisableCheckpoint {
		packer.CheckpointReporter = packer.NewCheckpointReporter(
			config.DisableCheckpointSignature,
		)
	}

	// Create the configuration for panicwrap and wrap our executable
	wrapConfig.Handler = panicHandler(logTempFile)
	wrapConfig.Writer = io.MultiWriter(logTempFile, &packersdk.LogSecretFilter)
	wrapConfig.Stdout = outW
	wrapConfig.DetectDuration = 500 * time.Millisecond
	wrapConfig.ForwardSignals = []os.Signal{syscall.SIGTERM}
	exitStatus, err := panicwrap.Wrap(&wrapConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't start Packer: %s", err)
		return 1
	}

	// If >= 0, we're the parent, so just exit
	if exitStatus >= 0 {
		// Close the stdout writer so that our copy process can finish
		outW.Close()

		// Wait for the output copying to finish
		<-doneCh

		return exitStatus
	}

	// We're the child, so just close the tempfile we made in order to
	// save file handles since the tempfile is only used by the parent.
	logTempFile.Close()

	return 0
}

// wrappedMain is called only when we're wrapped by panicwrap and
// returns the exit status to exit with.
func wrappedMain() int {
	// WARNING: WrappedMain causes unexpected behaviors when writing to stderr
	// and stdout.  Anything in this function written to stderr will be captured
	// by the logger, but will not be written to the terminal. Anything in
	// this function written to standard out must be prefixed with ErrorPrefix
	// or OutputPrefix to be sent to the right terminal stream, but adding
	// these prefixes can cause nondeterministic results for output from
	// other, asynchronous methods. Try to avoid modifying output in this
	// function if at all possible.

	// If there is no explicit number of Go threads to use, then set it
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	packersdk.LogSecretFilter.SetOutput(os.Stderr)
	log.SetOutput(&packersdk.LogSecretFilter)

	inPlugin := inPlugin()
	if inPlugin {
		// This prevents double-logging timestamps
		log.SetFlags(0)
	}

	log.Printf("[INFO] Packer version: %s [%s %s %s]",
		version.FormattedVersion(),
		runtime.Version(),
		runtime.GOOS, runtime.GOARCH)

	// The config being loaded here is the Packer config -- it defines
	// the location of third party builder plugins, plugin ports to use, and
	// whether to disable telemetry. It is a global config.
	// Do not confuse this config with the .json Packer template which gets
	// passed into commands like `packer build`
	config, err := loadConfig()
	if err != nil {
		// Writing to Stdout here so that the error message bypasses panicwrap. By using the
		// ErrorPrefix this output will be redirected to Stderr by the copyOutput func.
		// TODO: nywilken need to revisit this setup to better output errors to Stderr, and output to Stdout
		// without panicwrap
		fmt.Fprintf(os.Stdout, "%s Error loading configuration: \n\n%s\n", ErrorPrefix, err)
		return 1
	}

	// Fire off the checkpoint.
	go runCheckpoint(config)
	if !config.DisableCheckpoint {
		packer.CheckpointReporter = packer.NewCheckpointReporter(
			config.DisableCheckpointSignature,
		)
	}

	cacheDir, err := packersdk.CachePath()
	if err != nil {
		// Writing to Stdout here so that the error message bypasses panicwrap. By using the
		// ErrorPrefix this output will be redirected to Stderr by the copyOutput func.
		// TODO: nywilken need to revisit this setup to better output errors to Stderr, and output to Stdout
		// without panicwrap
		fmt.Fprintf(os.Stdout, "%s Error preparing cache directory: \n\n%s\n", ErrorPrefix, err)
		return 1
	}
	log.Printf("[INFO] Setting cache directory: %s", cacheDir)

	// Determine if we're in machine-readable mode by mucking around with
	// the arguments...
	args, machineReadable := extractMachineReadable(os.Args[1:])

	defer packer.CleanupClients()

	var ui packersdk.Ui
	if machineReadable {
		// Setup the UI as we're being machine-readable
		ui = &packer.MachineReadableUi{
			Writer: os.Stdout,
		}

		// Set this so that we don't get colored output in our machine-
		// readable UI.
		if err := os.Setenv("PACKER_NO_COLOR", "1"); err != nil {
			// Outputting error using Ui here to conform to the machine readable format.
			ui.Error(fmt.Sprintf("Packer failed to initialize UI: %s\n", err))
			return 1
		}
	} else {
		basicUi := &packersdk.BasicUi{
			Reader:      os.Stdin,
			Writer:      os.Stdout,
			ErrorWriter: os.Stdout,
			PB:          &packersdk.NoopProgressTracker{},
		}
		ui = basicUi
		if !inPlugin {
			currentPID := os.Getpid()
			backgrounded, err := checkProcess(currentPID)
			if err != nil {
				// Writing to Stderr will ensure that the output gets captured by panicwrap.
				// This error message and any other message writing to Stderr after this point will only show up with PACKER_LOG=1
				// TODO: nywilken need to revisit this setup to better output errors to Stderr, and output to Stdout without panicwrap.
				fmt.Fprintf(os.Stderr, "%s cannot determine if process is in background: %s\n", ErrorPrefix, err)
			}

			if backgrounded {
				fmt.Fprintf(os.Stderr, "%s Running in background, not using a TTY\n", ErrorPrefix)
			} else if TTY, err := openTTY(); err != nil {
				fmt.Fprintf(os.Stderr, "%s No tty available: %s\n", ErrorPrefix, err)
			} else {
				basicUi.TTY = TTY
				basicUi.PB = &packer.UiProgressBar{}
				defer TTY.Close()
			}
		}
	}
	// Create the CLI meta
	CommandMeta = &command.Meta{
		CoreConfig: &packer.CoreConfig{
			Components: packer.ComponentFinder{
				Hook:         config.StarHook,
				PluginConfig: config.Plugins,
			},
			Version: version.Version,
		},
		Ui: ui,
	}

	//versionCLIHelper shortcuts "--version" and "-v" to just show the version
	versionCLIHelper := &cli.CLI{
		Args:    args,
		Version: version.Version,
	}
	if versionCLIHelper.IsVersion() && versionCLIHelper.Version != "" {
		// by default version flags ignore all other args so there is no need to persist the original args.
		args = []string{"version"}
	}

	cli := &cli.CLI{
		Args:         args,
		Autocomplete: true,
		Commands:     Commands,
		HelpFunc:     excludeHelpFunc(Commands, []string{"execute", "plugin"}),
		HelpWriter:   os.Stdout,
		Name:         "packer",
		Version:      version.Version,
	}

	exitCode, err := cli.Run()
	if !inPlugin {
		if err := packer.CheckpointReporter.Finalize(cli.Subcommand(), exitCode, err); err != nil {
			log.Printf("[WARN] (telemetry) Error finalizing report. This is safe to ignore. %s", err.Error())
		}
	}

	if err != nil {
		// Writing to Stdout here so that the error message bypasses panicwrap. By using the
		// ErrorPrefix this output will be redirected to Stderr by the copyOutput func.
		// TODO: nywilken need to revisit this setup to better output errors to Stderr, and output to Stdout
		// without panicwrap
		fmt.Fprintf(os.Stdout, "%s Error executing CLI: %s\n", ErrorPrefix, err)
		return 1
	}

	return exitCode
}

// excludeHelpFunc filters commands we don't want to show from the list of
// commands displayed in packer's help text.
func excludeHelpFunc(commands map[string]cli.CommandFactory, exclude []string) cli.HelpFunc {
	// Make search slice into a map so we can use use the `if found` idiom
	// instead of a nested loop.
	var excludes = make(map[string]interface{}, len(exclude))
	for _, item := range exclude {
		excludes[item] = nil
	}

	// Create filtered list of commands
	helpCommands := []string{}
	for command := range commands {
		if _, found := excludes[command]; !found {
			helpCommands = append(helpCommands, command)
		}
	}

	return cli.FilteredHelpFunc(helpCommands, cli.BasicHelpFunc("packer"))
}

// extractMachineReadable checks the args for the machine readable
// flag and returns whether or not it is on. It modifies the args
// to remove this flag.
func extractMachineReadable(args []string) ([]string, bool) {
	for i, arg := range args {
		if arg == "-machine-readable" {
			// We found it. Slice it out.
			result := make([]string, len(args)-1)
			copy(result, args[:i])
			copy(result[i:], args[i+1:])
			return result, true
		}
	}

	return args, false
}

func loadConfig() (*config, error) {
	pluginDir, err := packer.PluginFolder()
	if err != nil {
		return nil, err
	}

	var config config
	config.Plugins = &packer.PluginConfig{
		PluginMinPort:   10000,
		PluginMaxPort:   25000,
		PluginDirectory: pluginDir,
		Builders:        packer.MapOfBuilder{},
		Provisioners:    packer.MapOfProvisioner{},
		PostProcessors:  packer.MapOfPostProcessor{},
		DataSources:     packer.MapOfDatasource{},
	}

	// Finally, try to use an internal plugin. Note that this will not override
	// any previously-loaded plugins.
	if err := config.discoverInternalComponents(); err != nil {
		return nil, err
	}

	// start by loading from PACKER_CONFIG if available
	configFilePath := os.Getenv("PACKER_CONFIG")
	if configFilePath == "" {
		var err error
		log.Print("[INFO] PACKER_CONFIG env var not set; checking the default config file path")
		configFilePath, err = pathing.ConfigFile()
		if err != nil {
			log.Printf("Error detecting default config file path: %s", err)
		}
	}
	if configFilePath == "" {
		return &config, nil
	}
	log.Printf("[INFO] PACKER_CONFIG env var set; attempting to open config file: %s", configFilePath)
	f, err := os.Open(configFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}

		log.Printf("[WARN] Config file doesn't exist: %s", configFilePath)
		return &config, nil
	}
	defer f.Close()

	// This loads a json config, defined in packer/config.go
	if err := decodeConfig(f, &config); err != nil {
		return nil, err
	}

	config.LoadExternalComponentsFromConfig()

	return &config, nil
}

// copyOutput uses output prefixes to determine whether data on stdout
// should go to stdout or stderr. This is due to panicwrap using stderr
// as the log and error channel.
func copyOutput(r io.Reader, doneCh chan<- struct{}) {
	defer close(doneCh)

	pr, err := prefixedio.NewReader(r)
	if err != nil {
		panic(err)
	}

	stderrR, err := pr.Prefix(ErrorPrefix)
	if err != nil {
		panic(err)
	}
	stdoutR, err := pr.Prefix(OutputPrefix)
	if err != nil {
		panic(err)
	}
	defaultR, err := pr.Prefix("")
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		io.Copy(os.Stderr, stderrR)
	}()
	go func() {
		defer wg.Done()
		io.Copy(os.Stdout, stdoutR)
	}()
	go func() {
		defer wg.Done()
		io.Copy(os.Stdout, defaultR)
	}()

	wg.Wait()
}

func inPlugin() bool {
	return os.Getenv(pluginsdk.MagicCookieKey) == pluginsdk.MagicCookieValue
}

func init() {
	// Seed the random number generator
	rand.Seed(time.Now().UTC().UnixNano())
}
