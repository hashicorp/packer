// The packer package contains the core components of Packer.
package packer

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
)

// The function type used to lookup Builder implementations.
type BuilderFunc func(name string) (Builder, error)

// The function type used to lookup Command implementations.
type CommandFunc func(name string) (Command, error)

// The function type used to lookup Hook implementations.
type HookFunc func(name string) (Hook, error)

// The function type used to lookup PostProcessor implementations.
type PostProcessorFunc func(name string) (PostProcessor, error)

// The function type used to lookup Provisioner implementations.
type ProvisionerFunc func(name string) (Provisioner, error)

// ComponentFinder is a struct that contains the various function
// pointers necessary to look up components of Packer such as builders,
// commands, etc.
type ComponentFinder struct {
	Builder       BuilderFunc
	Command       CommandFunc
	Hook          HookFunc
	PostProcessor PostProcessorFunc
	Provisioner   ProvisionerFunc
}

// The environment interface provides access to the configuration and
// state of a single Packer run.
//
// It allows for things such as executing CLI commands, getting the
// list of available builders, and more.
type Environment interface {
	Builder(string) (Builder, error)
	Cache() Cache
	Cli([]string) (int, error)
	Hook(string) (Hook, error)
	PostProcessor(string) (PostProcessor, error)
	Provisioner(string) (Provisioner, error)
	Ui() Ui
}

// An implementation of an Environment that represents the Packer core
// environment.
type coreEnvironment struct {
	cache      Cache
	commands   []string
	components ComponentFinder
	ui         Ui
}

// This struct configures new environments.
type EnvironmentConfig struct {
	Cache      Cache
	Commands   []string
	Components ComponentFinder
	Ui         Ui
}

type helpCommandEntry struct {
	i        int
	key      string
	synopsis string
}

// DefaultEnvironmentConfig returns a default EnvironmentConfig that can
// be used to create a new enviroment with NewEnvironment with sane defaults.
func DefaultEnvironmentConfig() *EnvironmentConfig {
	config := &EnvironmentConfig{}
	config.Commands = make([]string, 0)
	config.Ui = &BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stdout,
	}

	return config
}

// This creates a new environment
func NewEnvironment(config *EnvironmentConfig) (resultEnv Environment, err error) {
	if config == nil {
		err = errors.New("config must be given to initialize environment")
		return
	}

	env := &coreEnvironment{}
	env.cache = config.Cache
	env.commands = config.Commands
	env.components = config.Components
	env.ui = config.Ui

	// We want to make sure the components have valid function pointers.
	// If a function pointer was not given, we assume that the function
	// will just return a nil component.
	if env.components.Builder == nil {
		env.components.Builder = func(string) (Builder, error) { return nil, nil }
	}

	if env.components.Command == nil {
		env.components.Command = func(string) (Command, error) { return nil, nil }
	}

	if env.components.Hook == nil {
		env.components.Hook = func(string) (Hook, error) { return nil, nil }
	}

	if env.components.PostProcessor == nil {
		env.components.PostProcessor = func(string) (PostProcessor, error) { return nil, nil }
	}

	if env.components.Provisioner == nil {
		env.components.Provisioner = func(string) (Provisioner, error) { return nil, nil }
	}

	// The default cache is just the system temporary directory
	if env.cache == nil {
		env.cache = &FileCache{CacheDir: os.TempDir()}
	}

	resultEnv = env
	return
}

// Returns a builder of the given name that is registered with this
// environment.
func (e *coreEnvironment) Builder(name string) (b Builder, err error) {
	b, err = e.components.Builder(name)
	if err != nil {
		return
	}

	if b == nil {
		err = fmt.Errorf("No builder returned for name: %s", name)
	}

	return
}

// Returns the cache for this environment
func (e *coreEnvironment) Cache() Cache {
	return e.cache
}

// Returns a hook of the given name that is registered with this
// environment.
func (e *coreEnvironment) Hook(name string) (h Hook, err error) {
	h, err = e.components.Hook(name)
	if err != nil {
		return
	}

	if h == nil {
		err = fmt.Errorf("No hook returned for name: %s", name)
	}

	return
}

// Returns a PostProcessor for the given name that is registered with this
// environment.
func (e *coreEnvironment) PostProcessor(name string) (p PostProcessor, err error) {
	p, err = e.components.PostProcessor(name)
	if err != nil {
		return
	}

	if p == nil {
		err = fmt.Errorf("No post processor found for name: %s", name)
	}

	return
}

// Returns a provisioner for the given name that is registered with this
// environment.
func (e *coreEnvironment) Provisioner(name string) (p Provisioner, err error) {
	p, err = e.components.Provisioner(name)
	if err != nil {
		return
	}

	if p == nil {
		err = fmt.Errorf("No provisioner returned for name: %s", name)
	}

	return
}

// Executes a command as if it was typed on the command-line interface.
// The return value is the exit code of the command.
func (e *coreEnvironment) Cli(args []string) (result int, err error) {
	log.Printf("Environment.Cli: %#v\n", args)

	// If we have no arguments, just short-circuit here and print the help
	if len(args) == 0 {
		e.printHelp()
		return 1, nil
	}

	// This variable will track whether or not we're supposed to print
	// the help or not.
	isHelp := false
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			isHelp = true
			break
		}
	}

	// Trim up to the command name
	for i, v := range args {
		if len(v) > 0 && v[0] != '-' {
			args = args[i:]
			break
		}
	}

	log.Printf("command + args: %#v", args)

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
		command, err = e.components.Command(args[0])
		if err != nil {
			return
		}

		// If we still don't have a command, show the help.
		if command == nil {
			e.ui.Error(fmt.Sprintf("Unknown command: %s\n", args[0]))
			e.printHelp()
			return 1, nil
		}
	}

	// If we're supposed to print help, then print the help of the
	// command rather than running it.
	if isHelp {
		e.ui.Say(command.Help())
		return 0, nil
	}

	log.Printf("Executing command: %s\n", args[0])
	return command.Run(e, args[1:]), nil
}

// Prints the CLI help to the UI.
func (e *coreEnvironment) printHelp() {
	// Created a sorted slice of the map keys and record the longest
	// command name so we can better format the output later.
	maxKeyLen := 0
	for _, command := range e.commands {
		if len(command) > maxKeyLen {
			maxKeyLen = len(command)
		}
	}

	// Sort the keys
	sort.Strings(e.commands)

	// Create the communication/sync mechanisms to get the synopsis' of
	// the various commands. We do this in parallel since the overhead
	// of the subprocess underneath is very expensive and this speeds things
	// up an incredible amount.
	var wg sync.WaitGroup
	ch := make(chan *helpCommandEntry)

	for i, key := range e.commands {
		wg.Add(1)

		// Get the synopsis in a goroutine since it may take awhile
		// to subprocess out.
		go func(i int, key string) {
			defer wg.Done()
			var synopsis string
			command, err := e.components.Command(key)
			if err != nil {
				synopsis = fmt.Sprintf("Error loading command: %s", err.Error())
			} else if command == nil {
				return
			} else {
				synopsis = command.Synopsis()
			}

			// Pad the key with spaces so that they're all the same width
			key = fmt.Sprintf("%s%s", key, strings.Repeat(" ", maxKeyLen-len(key)))

			// Output the command and the synopsis
			ch <- &helpCommandEntry{
				i:        i,
				key:      key,
				synopsis: synopsis,
			}
		}(i, key)
	}

	e.ui.Say("usage: packer [--version] [--help] <command> [<args>]\n")
	e.ui.Say("Available commands are:")

	// Make a goroutine that just waits for all the synopsis gathering
	// to complete, and then output it.
	synopsisDone := make(chan struct{})
	go func() {
		defer close(synopsisDone)
		entries := make([]string, len(e.commands))

		for entry := range ch {
			e.ui.Machine("command", entry.key, entry.synopsis)
			message := fmt.Sprintf("    %s    %s", entry.key, entry.synopsis)
			entries[entry.i] = message
		}

		for _, message := range entries {
			if message != "" {
				e.ui.Say(message)
			}
		}
	}()

	// Wait to complete getting the synopsis' then close the channel
	wg.Wait()
	close(ch)
	<-synopsisDone

	e.ui.Say("\nGlobally recognized options:")
	e.ui.Say("    -machine-readable    Machine-readable output format.")
}

// Returns the UI for the environment. The UI is the interface that should
// be used for all communication with the outside world.
func (e *coreEnvironment) Ui() Ui {
	return e.ui
}
