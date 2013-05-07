// This is the main package for the `packer` application.
package main

import (
	"github.com/mitchellh/packer/packer"
	"fmt"
	"os"
)

func main() {
	envConfig := packer.DefaultEnvironmentConfig()

	env, err := packer.NewEnvironment(envConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Packer initialization error: \n\n%s\n", err)
		os.Exit(1)
	}

	os.Exit(env.Cli(os.Args[1:]))
}
