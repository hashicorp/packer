// This is the main package for the `packer` application.
package main

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/packer/plugin"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

func main() {
	// Setup logging if PACKER_LOG is set.
	// Log to PACKER_LOG_PATH if it is set, otherwise default to stderr.
	var logOutput io.Writer = ioutil.Discard
	if os.Getenv("PACKER_LOG") != "" {
		logOutput = os.Stderr

		if logPath := os.Getenv("PACKER_LOG_PATH"); logPath != "" {
			var err error
			logOutput, err = os.Create(logPath)
			if err != nil {
				fmt.Fprintf(
					os.Stderr,
					"Couldn't open '%s' for logging: %s",
					logPath, err)
				os.Exit(1)
			}
		}
	}

	log.SetOutput(logOutput)

	// If there is no explicit number of Go threads to use, then set it
	if os.Getenv("GOMAXPROCS") == "" {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	log.Printf(
		"Packer Version: %s %s %s",
		packer.Version, packer.VersionPrerelease, packer.GitCommit)
	log.Printf("Packer Target OS/Arch: %s %s", runtime.GOOS, runtime.GOARCH)

	config, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: \n\n%s\n", err)
		os.Exit(1)
	}

	log.Printf("Packer config: %+v", config)

	cacheDir := os.Getenv("PACKER_CACHE_DIR")
	if cacheDir == "" {
		cacheDir = "packer_cache"
	}

	cacheDir, err = filepath.Abs(cacheDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error preparing cache directory: \n\n%s\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error preparing cache directory: \n\n%s\n", err)
		os.Exit(1)
	}

	log.Printf("Setting cache directory: %s", cacheDir)
	cache := &packer.FileCache{CacheDir: cacheDir}

	defer plugin.CleanupClients()

	envConfig := packer.DefaultEnvironmentConfig()
	envConfig.Cache = cache
	envConfig.Commands = config.CommandNames()
	envConfig.Components.Builder = config.LoadBuilder
	envConfig.Components.Command = config.LoadCommand
	envConfig.Components.Hook = config.LoadHook
	envConfig.Components.PostProcessor = config.LoadPostProcessor
	envConfig.Components.Provisioner = config.LoadProvisioner

	env, err := packer.NewEnvironment(envConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Packer initialization error: \n\n%s\n", err)
		os.Exit(1)
	}

	setupSignalHandlers(env)

	exitCode, err := env.Cli(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())
		os.Exit(1)
	}

	plugin.CleanupClients()
	os.Exit(exitCode)
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
