// This is the main package for the `packer` application.
package main

import (
	"github.com/mitchellh/packer/packer"
	"os"
)

func main() {
	env := packer.NewEnvironment(nil)
	os.Exit(env.Cli(os.Args[1:]))
}
