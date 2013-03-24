// The packer package contains the core components of Packer.
package packer

import "os"

// A command is a runnable sub-command of the `packer` application.
// When `packer` is called with the proper subcommand, this will be
// called.
//
// The mapping of command names to command interfaces is in the
// Environment struct.
type Command interface {
	Run(env *Environment, args []string) int
}

// The environment struct contains all the state necessary for a single
// instance of Packer.
//
// It is *not* a singleton, but generally a single environment is created
// when Packer starts running to represent that Packer run. Technically,
// if you're building a custom Packer binary, you could instantiate multiple
// environments and run them in parallel.
type Environment struct {
	command map[string]Command
	ui      Ui
}

// This creates a new environment
func NewEnvironment() *Environment {
	env := &Environment{}
	env.command = make(map[string]Command)
	env.command["version"] = new(versionCommand)
	env.ui = &ReaderWriterUi{ os.Stdin, os.Stdout }
	return env
}

// Executes a command as if it was typed on the command-line interface.
// The return value is the exit code of the command.
func (e *Environment) Cli(args []string) int {
	if len(args) == 0 {
		e.PrintHelp()
		return 1
	}

	command, ok := e.command[args[0]]
	if !ok {
		// The command was not found. In this case, let's go through
		// the arguments and see if the user is requesting the version.
		for _, arg := range args {
			if arg == "--version" || arg == "-v" {
				command = e.command["version"]
				break
			}
		}

		// If we still don't have a command, show the help.
		if command == nil {
			e.PrintHelp()
			return 1
		}
	}

	return command.Run(e, args)
}

// Returns the UI for the environment. The UI is the interface that should
// be used for all communication with the outside world.
func (e *Environment) Ui() Ui {
	return e.ui
}

// Prints the CLI help to the UI.
func (e *Environment) PrintHelp() {
	e.ui.Say("Bad.\n")
}
