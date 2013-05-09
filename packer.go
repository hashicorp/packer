// This is the main package for the `packer` application.
package main

import (
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/packer/plugin"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	if os.Getenv("PACKER_LOG") == "" {
		// If we don't have logging explicitly enabled, then disable it
		log.SetOutput(ioutil.Discard)
	} else {
		// Logging is enabled, make sure it goes to stderr
		log.SetOutput(os.Stderr)
	}

	defer plugin.CleanupClients()

	config, err := parseConfig(defaultConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading global Packer configuration: \n\n%s\n", err)
		os.Exit(1)
	}

	envConfig := packer.DefaultEnvironmentConfig()
	envConfig.Commands = config.CommandNames()
	envConfig.CommandFunc = config.LoadCommand

	env, err := packer.NewEnvironment(envConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Packer initialization error: \n\n%s\n", err)
		os.Exit(1)
	}

	exitCode, err := env.Cli(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())
		os.Exit(1)
	}

	plugin.CleanupClients()
	os.Exit(exitCode)
}
