package main

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/packer/plugin"
	"log"
	"os/exec"
)

type config struct {
	builds map[string]string
	commands map[string]string
}

func defaultConfig() (result *config) {
	commands := []string{"build"}

	result = new(config)
	result.builds = make(map[string]string)
	result.commands = make(map[string]string)

	for _, name := range commands {
		result.commands[name] = fmt.Sprintf("packer-command-%s", name)
	}

	return
}

func (c *config) Commands() (result []string) {
	result = make([]string, 0, len(c.commands))
	for name, _ := range c.commands {
		result = append(result, name)
	}
	return
}

func (c *config) LoadCommand(name string) (packer.Command, error) {
	log.Printf("Loading command: %s\n", name)
	commandBin, ok := c.commands[name]
	if !ok {
		log.Printf("Command not found: %s\n", name)
		return nil, nil
	}

	return plugin.Command(exec.Command(commandBin))
}
