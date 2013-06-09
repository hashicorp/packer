package main

import (
	"encoding/json"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/packer/plugin"
	"io"
	"log"
	"os/exec"
)

// This is the default, built-in configuration that ships with
// Packer.
const defaultConfig = `
{
	"builders": {
		"amazon-ebs": "packer-builder-amazon-ebs",
		"vmware": "packer-builder-vmware"
	},

	"commands": {
		"build": "packer-command-build"
	},

	"provisioners": {
		"shell": "packer-provisioner-shell"
	}
}
`

type config struct {
	Builders     map[string]string
	Commands     map[string]string
	Provisioners map[string]string
}

// Decodes configuration in JSON format from the given io.Reader into
// the config object pointed to.
func decodeConfig(r io.Reader, c *config) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(c)
}

// Returns an array of defined command names.
func (c *config) CommandNames() (result []string) {
	result = make([]string, 0, len(c.Commands))
	for name, _ := range c.Commands {
		result = append(result, name)
	}
	return
}

// This is a proper packer.BuilderFunc that can be used to load packer.Builder
// implementations from the defined plugins.
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

// This is a proper implementation of packer.HookFunc that can be used
// to load packer.Hook implementations from the defined plugins.
func (c *config) LoadHook(name string) (packer.Hook, error) {
	log.Printf("Loading hook: %s\n", name)
	return plugin.Hook(exec.Command(name))
}

// This is a proper packer.ProvisionerFunc that can be used to load
// packer.Provisioner implementations from defined plugins.
func (c *config) LoadProvisioner(name string) (packer.Provisioner, error) {
	log.Printf("Loading provisioner: %s\n", name)
	provBin, ok := c.Provisioners[name]
	if !ok {
		log.Printf("Provisioner not found: %s\n", name)
		return nil, nil
	}

	return plugin.Provisioner(exec.Command(provBin))
}
