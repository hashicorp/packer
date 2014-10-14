package main

import (
	"encoding/json"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mitchellh/osext"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/packer/plugin"
)

type config struct {
	DisableCheckpoint          bool `json:"disable_checkpoint"`
	DisableCheckpointSignature bool `json:"disable_checkpoint_signature"`
	PluginMinPort              uint
	PluginMaxPort              uint

	Builders       map[string]string
	Commands       map[string]string
	PostProcessors map[string]string `json:"post-processors"`
	Provisioners   map[string]string
}

// ConfigFile returns the default path to the configuration file. On
// Unix-like systems this is the ".packerconfig" file in the home directory.
// On Windows, this is the "packer.config" file in the application data
// directory.
func ConfigFile() (string, error) {
	return configFile()
}

// ConfigDir returns the configuration directory for Packer.
func ConfigDir() (string, error) {
	return configDir()
}

// Decodes configuration in JSON format from the given io.Reader into
// the config object pointed to.
func decodeConfig(r io.Reader, c *config) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(c)
}

// Discover discovers plugins.
//
// This looks in the directory of the executable and the CWD, in that
// order for priority.
func (c *config) Discover() error {
	// Next, look in the same directory as the executable. Any conflicts
	// will overwrite those found in our current directory.
	exePath, err := osext.Executable()
	if err != nil {
		log.Printf("[ERR] Error loading exe directory: %s", err)
	} else {
		if err := c.discover(filepath.Dir(exePath)); err != nil {
			return err
		}
	}

	// Look in the plugins directory
	dir, err := ConfigDir()
	if err != nil {
		log.Printf("[ERR] Error loading config directory: %s", err)
	} else {
		if err := c.discover(filepath.Join(dir, "plugins")); err != nil {
			return err
		}
	}

	// Look in the cwd.
	if err := c.discover("."); err != nil {
		return err
	}

	return nil
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

func (c *config) discover(path string) error {
	var err error

	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil {
			return err
		}
	}

	err = c.discoverSingle(
		filepath.Join(path, "packer-builder-*"), &c.Builders)
	if err != nil {
		return err
	}

	err = c.discoverSingle(
		filepath.Join(path, "packer-command-*"), &c.Commands)
	if err != nil {
		return err
	}

	err = c.discoverSingle(
		filepath.Join(path, "packer-post-processor-*"), &c.PostProcessors)
	if err != nil {
		return err
	}

	err = c.discoverSingle(
		filepath.Join(path, "packer-provisioner-*"), &c.Provisioners)
	if err != nil {
		return err
	}

	return nil
}

func (c *config) discoverSingle(glob string, m *map[string]string) error {
	matches, err := filepath.Glob(glob)
	if err != nil {
		return err
	}

	if *m == nil {
		*m = make(map[string]string)
	}

	prefix := filepath.Base(glob)
	prefix = prefix[:strings.Index(prefix, "*")]
	for _, match := range matches {
		file := filepath.Base(match)

		// If the filename has a ".", trim up to there
		if idx := strings.Index(file, "."); idx >= 0 {
			file = file[:idx]
		}

		// Look for foo-bar-baz. The plugin name is "baz"
		plugin := file[len(prefix):]
		log.Printf("[DEBUG] Discoverd plugin: %s = %s", plugin, match)
		(*m)[plugin] = match
	}

	return nil
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
