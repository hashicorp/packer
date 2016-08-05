// This is the main package for the `packer` application.

//go:generate go run ./scripts/generate-plugins.go
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/hashicorp/go-uuid"
	"github.com/mitchellh/cli"
	"github.com/mitchellh/packer/command"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/mitchellh/packer/version"
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

	if !panicwrap.Wrapped(&wrapConfig) {
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

		// We always send logs to a temporary file that we use in case
		// there is a panic. Otherwise, we delete it.
		logTempFile, err := ioutil.TempFile("", "packer-log")
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

		// Create the configuration for panicwrap and wrap our executable
		wrapConfig.Handler = panicHandler(logTempFile)
		wrapConfig.Writer = io.MultiWriter(logTempFile, logWriter)
		wrapConfig.Stdout = outW
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
	}

	// Call the real main
	return wrappedMain()
}

// wrappedMain is called only when we're wrapped by panicwrap and
// returns the exit status to exit with.
func wrappedMain() int {
	// If there is no explicit number of Go threads to use, then set it
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	log.SetOutput(os.Stderr)

	log.Printf("[INFO] Packer version: %s", version.FormattedVersion())
	log.Printf("Packer Target OS/Arch: %s %s", runtime.GOOS, runtime.GOARCH)
	log.Printf("Built with Go Version: %s", runtime.Version())

	// Prepare stdin for plugin usage by switching it to a pipe
	// But do not switch to pipe in plugin
	if os.Getenv(plugin.MagicCookieKey) != plugin.MagicCookieValue {
		setupStdin()
	}

	config, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: \n\n%s\n", err)
		return 1
	}
	log.Printf("Packer config: %+v", config)

	// Fire off the checkpoint.
	go runCheckpoint(config)

	cacheDir := os.Getenv("PACKER_CACHE_DIR")
	if cacheDir == "" {
		cacheDir = "packer_cache"
	}

	cacheDir, err = filepath.Abs(cacheDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error preparing cache directory: \n\n%s\n", err)
		return 1
	}

	log.Printf("Setting cache directory: %s", cacheDir)
	cache := &packer.FileCache{CacheDir: cacheDir}

	// Determine if we're in machine-readable mode by mucking around with
	// the arguments...
	args, machineReadable := extractMachineReadable(os.Args[1:])

	defer plugin.CleanupClients()

	// Setup the UI if we're being machine-readable
	var ui packer.Ui = &packer.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stdout,
	}
	if machineReadable {
		ui = &packer.MachineReadableUi{
			Writer: os.Stdout,
		}

		// Set this so that we don't get colored output in our machine-
		// readable UI.
		if err := os.Setenv("PACKER_NO_COLOR", "1"); err != nil {
			fmt.Fprintf(os.Stderr, "Packer failed to initialize UI: %s\n", err)
			return 1
		}
	}

	// Create the CLI meta
	CommandMeta = &command.Meta{
		CoreConfig: &packer.CoreConfig{
			Components: packer.ComponentFinder{
				Builder:       config.LoadBuilder,
				Hook:          config.LoadHook,
				PostProcessor: config.LoadPostProcessor,
				Provisioner:   config.LoadProvisioner,
			},
			Version: version.Version,
		},
		Cache: cache,
		Ui:    ui,
	}

	//setupSignalHandlers(env)

	cli := &cli.CLI{
		Args:       args,
		Commands:   Commands,
		HelpFunc:   excludeHelpFunc(Commands, []string{"plugin"}),
		HelpWriter: os.Stdout,
		Version:    version.Version,
	}

	exitCode, err := cli.Run()
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
	if err := config.Discover(); err != nil {
		return nil, err
	}

	configFilePath := os.Getenv("PACKER_CONFIG")
	if configFilePath == "" {
		var err error
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

	if err := decodeConfig(f, &config); err != nil {
		return nil, err
	}

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

func init() {
	// Seed the random number generator
	rand.Seed(time.Now().UTC().UnixNano())
}
