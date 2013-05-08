// This is the main package for the `packer` application.
package main

import (
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/packer/plugin"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	commands := map[string]string {
		"build": "packer-build",
	}

	commandKeys := make([]string, 0, len(commands))
	for k, _ := range commands {
		commandKeys = append(commandKeys, k)
	}

	envConfig := packer.DefaultEnvironmentConfig()
	envConfig.Commands = commandKeys
	envConfig.CommandFunc = func(n string) (packer.Command, error) {
		commandBin, ok := commands[n]
		if !ok {
			return nil, nil
		}

		return plugin.Command(exec.Command(commandBin))
	}

	env, err := packer.NewEnvironment(envConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Packer initialization error: \n\n%s\n", err)
		os.Exit(1)
	}

	exitCode, _ := env.Cli(os.Args[1:])
	os.Exit(exitCode)
}
