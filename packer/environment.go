// The packer package contains the core components of Packer.
package packer

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

type BuilderFunc func(name string) Builder

type CommandFunc func(name string) Command

// The environment interface provides access to the configuration and
// state of a single Packer run.
//
// It allows for things such as executing CLI commands, getting the
// list of available builders, and more.
type Environment interface {
	Builder(name string) Builder
	Cli(args []string) int
	Ui() Ui
}

// An implementation of an Environment that represents the Packer core
// environment.
type coreEnvironment struct {
	builderFunc BuilderFunc
	commands []string
	commandFunc CommandFunc
	ui      Ui
}

// This struct configures new environments.
type EnvironmentConfig struct {
	BuilderFunc BuilderFunc
	CommandFunc CommandFunc
	Commands []string
	Ui      Ui
}

// DefaultEnvironmentConfig returns a default EnvironmentConfig that can
// be used to create a new enviroment with NewEnvironment with sane defaults.
func DefaultEnvironmentConfig() *EnvironmentConfig {
	config := &EnvironmentConfig{}
	config.BuilderFunc = func(string) Builder { return nil }
	config.CommandFunc = func(string) Command { return nil }
	config.Commands = make([]string, 0)
	config.Ui = &ReaderWriterUi{os.Stdin, os.Stdout}
	return config
}

// This creates a new environment
func NewEnvironment(config *EnvironmentConfig) (resultEnv Environment, err error) {
	if config == nil {
		err = errors.New("config must be given to initialize environment")
		return
	}

	env := &coreEnvironment{}
	env.builderFunc = config.BuilderFunc
	env.commandFunc = config.CommandFunc
	env.commands = config.Commands
	env.ui = config.Ui

	resultEnv = env
	return
}

// Returns a builder of the given name that is registered with this
// environment.
func (e *coreEnvironment) Builder(name string) Builder {
	return e.builderFunc(name)
}

// Executes a command as if it was typed on the command-line interface.
// The return value is the exit code of the command.
func (e *coreEnvironment) Cli(args []string) int {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		e.printHelp()
		return 1
	}

	version := args[0] == "version"
	if !version {
		for _, arg := range args {
			if arg == "--version" || arg == "-v" {
				version = true
				break
			}
		}
	}

	var command Command
	if version {
		command = new(versionCommand)
	}

	if command == nil {
		command = e.commandFunc(args[0])

		// If we still don't have a command, show the help.
		if command == nil {
			e.printHelp()
			return 1
		}
	}

	return command.Run(e, args[1:])
}

// Prints the CLI help to the UI.
func (e *coreEnvironment) printHelp() {
	// Created a sorted slice of the map keys and record the longest
	// command name so we can better format the output later.
	i := 0
	maxKeyLen := 0
	for _, command := range e.commands {
		if len(command) > maxKeyLen {
			maxKeyLen = len(command)
		}

		i++
	}

	// Sort the keys
	sort.Strings(e.commands)

	e.ui.Say("usage: packer [--version] [--help] <command> [<args>]\n\n")
	e.ui.Say("Available commands are:\n")
	for _, key := range e.commands {
		command := e.commandFunc(key)

		// Pad the key with spaces so that they're all the same width
		key = fmt.Sprintf("%v%v", key, strings.Repeat(" ", maxKeyLen-len(key)))

		// Output the command and the synopsis
		e.ui.Say("    %v     %v\n", key, command.Synopsis())
	}
}

// Returns the UI for the environment. The UI is the interface that should
// be used for all communication with the outside world.
func (e *coreEnvironment) Ui() Ui {
	return e.ui
}
