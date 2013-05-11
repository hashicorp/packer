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
[builders]
amazon-ebs = "packer-builder-amazon-ebs"

[commands]
build = "packer-command-build"
`

type config struct {
	Builders map[string]string
	Commands map[string]string
}

// Merge the configurations. Anything in the "new" configuration takes
// precedence over the "old" configuration.
func mergeConfig(a, b *config) *config {
	configs := []*config{a, b}
	result := newConfig()

	for _, config := range configs {
		for k, v := range config.Builders {
			result.Builders[k] = v
		}

		for k, v := range config.Commands {
			result.Commands[k] = v
		}
	}

	return result
}

// Creates and initializes a new config struct.
func newConfig() *config {
	result := new(config)
	result.Builders = make(map[string]string)
	result.Commands = make(map[string]string)
	return result
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

func (c *config) LoadBuilder(name string) (packer.Builder, error) {
	log.Printf("Loading builder: %s\n", name)
	bin, ok := c.Builders[name]
	if !ok {
		log.Printf("Builder not found: %s\n", name)
		return nil, nil
	}

	return plugin.Builder(exec.Command(bin))
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

func (c *config) LoadHook(name string) (packer.Hook, error) {
	log.Printf("Loading hook: %s\n", name)
	return plugin.Hook(exec.Command(name))
}
