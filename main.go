// This is the main package for the `packer` application.

//go:generate go run ./scripts/generate-plugins.go
//go:generate go generate ./packer-plugin-sdk/bootcommand/...
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/packer/command"
	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/tmp"
	"github.com/hashicorp/packer/packer/plugin"
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
		logWriter = ioutil.Discard
	}

	packer.LogSecretFilter.SetOutput(logWriter)

	// Disable logging here
	log.SetOutput(ioutil.Discard)

	// We always send logs to a temporary file that we use in case
	// there is a panic. Otherwise, we delete it.
	logTempFile, err := tmp.File("packer-log")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't setup logging tempfile: %s", err)
		return 1
	}
	defer os.Remove(logTempFile.Name())
	defer logTempFile.Close()

	// Tell the logger to log to this file
	os.Setenv(EnvLog, "")
	os.Setenv(EnvLogFile, "")

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
	wrapConfig.Writer = io.MultiWriter(logTempFile, &packer.LogSecretFilter)
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
	// If there is no explicit number of Go threads to use, then set it
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	packer.LogSecretFilter.SetOutput(os.Stderr)
	log.SetOutput(&packer.LogSecretFilter)

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
		fmt.Fprintf(os.Stderr, "Error loading configuration: \n\n%s\n", err)
		return 1
	}

	// Fire off the checkpoint.
	go runCheckpoint(config)
	if !config.DisableCheckpoint {
		packer.CheckpointReporter = packer.NewCheckpointReporter(
			config.DisableCheckpointSignature,
		)
	}

	cacheDir, err := packer.CachePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error preparing cache directory: \n\n%s\n", err)
		return 1
	}
	log.Printf("Setting cache directory: %s", cacheDir)

	// Determine if we're in machine-readable mode by mucking around with
	// the arguments...
	args, machineReadable := extractMachineReadable(os.Args[1:])

	defer plugin.CleanupClients()

	var ui packersdk.Ui
	if machineReadable {
		// Setup the UI as we're being machine-readable
		ui = &packer.MachineReadableUi{
			Writer: os.Stdout,
		}

		// Set this so that we don't get colored output in our machine-
		// readable UI.
		if err := os.Setenv("PACKER_NO_COLOR", "1"); err != nil {
			fmt.Fprintf(os.Stderr, "Packer failed to initialize UI: %s\n", err)
			return 1
		}
	} else {
		basicUi := &packer.BasicUi{
			Reader:      os.Stdin,
			Writer:      os.Stdout,
			ErrorWriter: os.Stdout,
			PB:          &packer.NoopProgressTracker{},
		}
		ui = basicUi
		if !inPlugin {
			currentPID := os.Getpid()
			backgrounded, err := checkProcess(currentPID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "cannot determine if process is in "+
					"background: %s\n", err)
			}
			if backgrounded {
				fmt.Fprint(os.Stderr, "Running in background, not using a TTY\n")
			} else if TTY, err := openTTY(); err != nil {
				fmt.Fprintf(os.Stderr, "No tty available: %s\n", err)
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
				Hook: config.StarHook,

				BuilderStore:       config.Builders,
				ProvisionerStore:   config.Provisioners,
				PostProcessorStore: config.PostProcessors,
			},
			Version: version.Version,
		},
		Ui: ui,
	}

	cli := &cli.CLI{
		Args:         args,
		Autocomplete: true,
		Commands:     Commands,
		HelpFunc:     excludeHelpFunc(Commands, []string{"plugin"}),
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
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err)
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
	var config config
	config.PluginMinPort = 10000
	config.PluginMaxPort = 25000
	config.Builders = packer.MapOfBuilder{}
	config.PostProcessors = packer.MapOfPostProcessor{}
	config.Provisioners = packer.MapOfProvisioner{}
	if err := config.Discover(); err != nil {
		return nil, err
	}

	// start by loading from PACKER_CONFIG if available
	log.Print("Checking 'PACKER_CONFIG' for a config file path")
	configFilePath := os.Getenv("PACKER_CONFIG")

	if configFilePath == "" {
		var err error
		log.Print("'PACKER_CONFIG' not set; checking the default config file path")
		configFilePath, err = packer.ConfigFile()
		if err != nil {
			log.Printf("Error detecting default config file path: %s", err)
		}
	}

	if configFilePath == "" {
		return &config, nil
	}

	log.Printf("Attempting to open config file: %s", configFilePath)
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
	return os.Getenv(plugin.MagicCookieKey) == plugin.MagicCookieValue
}

func init() {
	// Seed the random number generator
	rand.Seed(time.Now().UTC().UnixNano())
}
