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
	envConfig := packer.DefaultEnvironmentConfig()
	envConfig.Commands = []string{"build"}
	envConfig.CommandFunc = func(n string) packer.Command {
		return plugin.Command(exec.Command("bin/packer-build"))
	}

	env, err := packer.NewEnvironment(envConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Packer initialization error: \n\n%s\n", err)
		os.Exit(1)
	}

	os.Exit(env.Cli(os.Args[1:]))
}
