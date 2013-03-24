// This is the main package for the `packer` application.
package main

import "github.com/mitchellh/packer/builder/amazon"

// A command is a runnable sub-command of the `packer` application.
// When `packer` is called with the proper subcommand, this will be
// called.
//
// The mapping of command names to command interfaces is in the
// Environment struct.
type Command interface {
	Run(args []string)
}

// The environment struct contains all the state necessary for a single
// instance of Packer.
//
// It is *not* a singleton, but generally a single environment is created
// when Packer starts running to represent that Packer run. Technically,
// if you're building a custom Packer binary, you could instantiate multiple
// environments and run them in parallel.
type Environment struct {
	commands map[string]Command
}

type Template struct {
	Name         string
	Builders     map[string]interface{} `toml:"builder"`
	Provisioners map[string]interface{} `toml:"provision"`
	Outputs      map[string]interface{} `toml:"output"`
}

type Builder interface {
	Prepare()
	Build()
	Destroy()
}

func main() {
	var builder Builder
	builder = &amazon.Builder{}
	builder.Build()
}
