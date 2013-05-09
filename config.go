package main

import (
	"github.com/BurntSushi/toml"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/packer/plugin"
	"log"
	"os/exec"
)

// This is the default, built-in configuration that ships with
// Packer.
const defaultConfig = `
[commands]
build = "packer-command-build"
`

type config struct {
	Builders map[string]string
	Commands map[string]string
}

// Parses a configuration file and returns a proper configuration
// struct.
func parseConfig(data string) (result *config, err error) {
	result = new(config)
	_, err = toml.Decode(data, &result)
	return
}

// Returns an array of defined command names.
func (c *config) CommandNames() (result []string) {
	result = make([]string, 0, len(c.Commands))
	for name, _ := range c.Commands {
		result = append(result, name)
	}
	return
}

// This is a proper packer.CommandFunc that can be used to load packer.Command
// implementations from the defined plugins.
func (c *config) LoadCommand(name string) (packer.Command, error) {
	log.Printf("Loading command: %s\n", name)
	commandBin, ok := c.Commands[name]
	if !ok {
		log.Printf("Command not found: %s\n", name)
		return nil, nil
	}

	return plugin.Command(exec.Command(commandBin))
}
