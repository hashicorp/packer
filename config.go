package main

import (
	"encoding/json"
	"github.com/mitchellh/osext"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/packer/plugin"
	"io"
	"log"
	"os/exec"
	"path/filepath"
)

// This is the default, built-in configuration that ships with
// Packer.
const defaultConfig = `
{
	"plugin_min_port": 10000,
	"plugin_max_port": 25000,

	"builders": {
		"amazon-ebs": "packer-builder-amazon-ebs",
		"digitalocean": "packer-builder-digitalocean",
		"virtualbox": "packer-builder-virtualbox",
		"vmware": "packer-builder-vmware"
	},

	"commands": {
		"build": "packer-command-build",
		"fix": "packer-command-fix",
		"validate": "packer-command-validate"
	},

	"post-processors": {
		"vagrant": "packer-post-processor-vagrant"
	},

	"provisioners": {
		"file": "packer-provisioner-file",
		"shell": "packer-provisioner-shell"
	}
}
`

type config struct {
	PluginMinPort uint
	PluginMaxPort uint

	Builders       map[string]string
	Commands       map[string]string
	PostProcessors map[string]string `json:"post-processors"`
	Provisioners   map[string]string
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
	for name := range c.Commands {
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

	return c.pluginClient(bin).Builder()
}

// This is a proper packer.CommandFunc that can be used to load packer.Command
// implementations from the defined plugins.
func (c *config) LoadCommand(name string) (packer.Command, error) {
	log.Printf("Loading command: %s\n", name)
	bin, ok := c.Commands[name]
	if !ok {
		log.Printf("Command not found: %s\n", name)
		return nil, nil
	}

	return c.pluginClient(bin).Command()
}

// This is a proper implementation of packer.HookFunc that can be used
// to load packer.Hook implementations from the defined plugins.
func (c *config) LoadHook(name string) (packer.Hook, error) {
	log.Printf("Loading hook: %s\n", name)
	return c.pluginClient(name).Hook()
}

// This is a proper packer.PostProcessorFunc that can be used to load
// packer.PostProcessor implementations from defined plugins.
func (c *config) LoadPostProcessor(name string) (packer.PostProcessor, error) {
	log.Printf("Loading post-processor: %s", name)
	bin, ok := c.PostProcessors[name]
	if !ok {
		log.Printf("Post-processor not found: %s", name)
		return nil, nil
	}

	return c.pluginClient(bin).PostProcessor()
}

// This is a proper packer.ProvisionerFunc that can be used to load
// packer.Provisioner implementations from defined plugins.
func (c *config) LoadProvisioner(name string) (packer.Provisioner, error) {
	log.Printf("Loading provisioner: %s\n", name)
	bin, ok := c.Provisioners[name]
	if !ok {
		log.Printf("Provisioner not found: %s\n", name)
		return nil, nil
	}

	return c.pluginClient(bin).Provisioner()
}

func (c *config) pluginClient(path string) *plugin.Client {
	originalPath := path

	// First attempt to find the executable by consulting the PATH.
	path, err := exec.LookPath(path)
	if err != nil {
		// If that doesn't work, look for it in the same directory
		// as the `packer` executable (us).
		log.Printf("Plugin could not be found. Checking same directory as executable.")
		exePath, err := osext.Executable()
		if err != nil {
			log.Printf("Couldn't get current exe path: %s", err)
		} else {
			log.Printf("Current exe path: %s", exePath)
			path = filepath.Join(filepath.Dir(exePath), filepath.Base(originalPath))
		}
	}

	// If everything failed, just use the original path and let the error
	// bubble through.
	if path == "" {
		path = originalPath
	}

	log.Printf("Creating plugin client for path: %s", path)
	var config plugin.ClientConfig
	config.Cmd = exec.Command(path)
	config.Managed = true
	config.MinPort = c.PluginMinPort
	config.MaxPort = c.PluginMaxPort
	return plugin.NewClient(&config)
}
