// This is the main package for the `packer` application.
package main

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/packer/plugin"
	"github.com/mitchellh/panicwrap"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	// Call realMain instead of doing the work here so we can use
	// `defer` statements within the function and have them work properly.
	// (defers aren't called with os.Exit)
	os.Exit(realMain())
}

// realMain is executed from main and returns the exit status to exit with.
func realMain() int {
	// If there is no explicit number of Go threads to use, then set it
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	// Determine where logs should go in general (requested by the user)
	logWriter, err := logOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't setup log output: %s", err)
		return 1
	}

	// We also always send logs to a temporary file that we use in case
	// there is a panic. Otherwise, we delete it.
	logTempFile, err := ioutil.TempFile("", "packer-log")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't setup logging tempfile: %s", err)
		return 1
	}
	defer os.Remove(logTempFile.Name())
	defer logTempFile.Close()

	// Reset the log variables to minimize work in the subprocess
	os.Setenv("PACKER_LOG", "")
	os.Setenv("PACKER_LOG_FILE", "")

	// Create the configuration for panicwrap and wrap our executable
	wrapConfig := &panicwrap.WrapConfig{
		Handler: panicHandler(logTempFile),
		Writer:  io.MultiWriter(logTempFile, logWriter),
	}

	exitStatus, err := panicwrap.Wrap(wrapConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't start Packer: %s", err)
		return 1
	}

	if exitStatus >= 0 {
		return exitStatus
	}

	// We're the child, so just close the tempfile we made in order to
	// save file handles since the tempfile is only used by the parent.
	logTempFile.Close()

	return wrappedMain()
}

// wrappedMain is called only when we're wrapped by panicwrap and
// returns the exit status to exit with.
func wrappedMain() int {
	log.SetOutput(os.Stderr)

	log.Printf(
		"Packer Version: %s %s %s",
		packer.Version, packer.VersionPrerelease, packer.GitCommit)
	log.Printf("Packer Target OS/Arch: %s %s", runtime.GOOS, runtime.GOARCH)
	log.Printf("Built with Go Version: %s", runtime.Version())

	// Prepare stdin for plugin usage by switching it to a pipe
	setupStdin()

	config, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: \n\n%s\n", err)
		return 1
	}

	log.Printf("Packer config: %+v", config)

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

	// Create the environment configuration
	envConfig := packer.DefaultEnvironmentConfig()
	envConfig.Cache = cache
	envConfig.Commands = config.CommandNames()
	envConfig.Components.Builder = config.LoadBuilder
	envConfig.Components.Command = config.LoadCommand
	envConfig.Components.Hook = config.LoadHook
	envConfig.Components.PostProcessor = config.LoadPostProcessor
	envConfig.Components.Provisioner = config.LoadProvisioner
	if machineReadable {
		envConfig.Ui = &packer.MachineReadableUi{
			Writer: os.Stdout,
		}

		// Set this so that we don't get colored output in our machine-
		// readable UI.
		if err := os.Setenv("PACKER_NO_COLOR", "1"); err != nil {
			fmt.Fprintf(os.Stderr, "Packer failed to initialize UI: %s\n", err)
			return 1
		}
	}

	env, err := packer.NewEnvironment(envConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Packer initialization error: \n\n%s\n", err)
		return 1
	}

	setupSignalHandlers(env)

	exitCode, err := env.Cli(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())
		return 1
	}

	return exitCode
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
	if err := decodeConfig(bytes.NewBufferString(defaultConfig), &config); err != nil {
		return nil, err
	}

	mustExist := true
	configFilePath := os.Getenv("PACKER_CONFIG")
	if configFilePath == "" {
		var err error
		configFilePath, err = configFile()
		mustExist = false

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

		if mustExist {
			return nil, err
		}

		log.Println("File doesn't exist, but doesn't need to. Ignoring.")
		return &config, nil
	}
	defer f.Close()

	if err := decodeConfig(f, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// logOutput determines where we should send logs (if anywhere).
func logOutput() (logOutput io.Writer, err error) {
	logOutput = ioutil.Discard
	if os.Getenv("PACKER_LOG") != "" {
		logOutput = os.Stderr

		if logPath := os.Getenv("PACKER_LOG_PATH"); logPath != "" {
			var err error
			logOutput, err = os.Create(logPath)
			if err != nil {
				return nil, err
			}
		}
	}

	return
}
