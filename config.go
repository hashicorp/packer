package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/hashicorp/packer/command"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer/plugin"
)

// PACKERSPACE is used to represent the spaces that separate args for a command
// without being confused with spaces in the path to the command itself.
const PACKERSPACE = "-PACKERSPACE-"

type config struct {
	DisableCheckpoint          bool                         `json:"disable_checkpoint"`
	DisableCheckpointSignature bool                         `json:"disable_checkpoint_signature"`
	RawBuilders                map[string]string            `json:"builders"`
	RawProvisioners            map[string]string            `json:"provisioners"`
	RawPostProcessors          map[string]string            `json:"post-processors"`
	Builders                   packersdk.MapOfBuilder       `json:"-"`
	Provisioners               packersdk.MapOfProvisioner   `json:"-"`
	PostProcessors             packersdk.MapOfPostProcessor `json:"-"`
	Plugins                    plugin.Config
}

// decodeConfig decodes configuration in JSON format from the given io.Reader into
// the config object pointed to.
func decodeConfig(r io.Reader, c *config) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(c)
}

// LoadExternalComponentsFromConfig loads plugins defined in RawBuilders, RawProvisioners, and RawPostProcessors.
func (c *config) LoadExternalComponentsFromConfig() {
	// helper to build up list of plugin paths
	extractPaths := func(m map[string]string) []string {
		paths := make([]string, 0, len(m))
		for _, v := range m {
			paths = append(paths, v)
		}

		return paths
	}

	var pluginPaths []string
	pluginPaths = append(pluginPaths, extractPaths(c.RawProvisioners)...)
	pluginPaths = append(pluginPaths, extractPaths(c.RawBuilders)...)
	pluginPaths = append(pluginPaths, extractPaths(c.RawPostProcessors)...)

	var externallyUsed = make([]string, 0, len(pluginPaths))
	for _, pluginPath := range pluginPaths {
		name, err := c.loadSingleComponent(pluginPath)
		if err != nil {
			log.Print(err)
			continue
		}

		log.Printf("loaded plugin: %s = %s", name, pluginPath)
		externallyUsed = append(externallyUsed, name)
	}

	if len(externallyUsed) > 0 {
		sort.Strings(externallyUsed)
		log.Printf("using external plugins %v", externallyUsed)
	}
}

func (c *config) loadSingleComponent(path string) (string, error) {
	pluginName := filepath.Base(path)

	// On Windows, ignore any plugins that don't end in .exe.
	// We could do a full PATHEXT parse, but this is probably good enough.
	if runtime.GOOS == "windows" && strings.ToLower(filepath.Ext(pluginName)) != ".exe" {
		return "", fmt.Errorf("error loading plugin %q, no exe extension", path)
	}

	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("error loading plugin %q: %s", path, err)
	}

	// If the filename has a ".", trim up to there
	if idx := strings.Index(pluginName, "."); idx >= 0 {
		pluginName = pluginName[:idx]
	}

	switch {
	case strings.HasPrefix(pluginName, "packer-builder-"):
		pluginName = pluginName[len("packer-builder-"):]
		c.Builders[pluginName] = func() (packersdk.Builder, error) {
			return c.Plugins.Client(path).Builder()
		}
	case strings.HasPrefix(pluginName, "packer-post-processor-"):
		pluginName = pluginName[len("packer-post-processor-"):]
		c.PostProcessors[pluginName] = func() (packersdk.PostProcessor, error) {
			return c.Plugins.Client(path).PostProcessor()
		}
	case strings.HasPrefix(pluginName, "packer-provisioner-"):
		pluginName = pluginName[len("packer-provisioner-"):]
		c.Provisioners[pluginName] = func() (packersdk.Provisioner, error) {
			return c.Plugins.Client(path).Provisioner()
		}
	}

	return pluginName, nil
}

// This is a proper packer.BuilderFunc that can be used to load packersdk.Builder
// implementations from the defined plugins.
func (c *config) StartBuilder(name string) (packersdk.Builder, error) {
	log.Printf("Loading builder: %s\n", name)
	return c.Builders.Start(name)
}

// This is a proper implementation of packer.HookFunc that can be used
// to load packersdk.Hook implementations from the defined plugins.
func (c *config) StarHook(name string) (packersdk.Hook, error) {
	log.Printf("Loading hook: %s\n", name)
	return c.Plugins.Client(name).Hook()
}

// This is a proper packersdk.PostProcessorFunc that can be used to load
// packersdk.PostProcessor implementations from defined plugins.
func (c *config) StartPostProcessor(name string) (packersdk.PostProcessor, error) {
	log.Printf("Loading post-processor: %s", name)
	return c.PostProcessors.Start(name)
}

// This is a proper packer.ProvisionerFunc that can be used to load
// packer.Provisioner implementations from defined plugins.
func (c *config) StartProvisioner(name string) (packersdk.Provisioner, error) {
	log.Printf("Loading provisioner: %s\n", name)
	return c.Provisioners.Start(name)
}

func (c *config) discoverExternalComponents(path string) error {
	var err error

	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil {
			return err
		}
	}
	externallyUsed := []string{}

	pluginPaths, err := c.discoverSingle(filepath.Join(path, "packer-builder-*"))
	if err != nil {
		return err
	}
	for pluginName, pluginPath := range pluginPaths {
		newPath := pluginPath // this needs to be stored in a new variable for the func below
		c.Builders[pluginName] = func() (packersdk.Builder, error) {
			return c.Plugins.Client(newPath).Builder()
		}
		externallyUsed = append(externallyUsed, pluginName)
	}
	if len(externallyUsed) > 0 {
		sort.Strings(externallyUsed)
		log.Printf("using external builders %v", externallyUsed)
		externallyUsed = nil
	}

	pluginPaths, err = c.discoverSingle(filepath.Join(path, "packer-post-processor-*"))
	if err != nil {
		return err
	}
	for pluginName, pluginPath := range pluginPaths {
		newPath := pluginPath // this needs to be stored in a new variable for the func below
		c.PostProcessors[pluginName] = func() (packersdk.PostProcessor, error) {
			return c.Plugins.Client(newPath).PostProcessor()
		}
		externallyUsed = append(externallyUsed, pluginName)
	}
	if len(externallyUsed) > 0 {
		sort.Strings(externallyUsed)
		log.Printf("using external post-processors %v", externallyUsed)
		externallyUsed = nil
	}

	pluginPaths, err = c.discoverSingle(filepath.Join(path, "packer-provisioner-*"))
	if err != nil {
		return err
	}
	for pluginName, pluginPath := range pluginPaths {
		newPath := pluginPath // this needs to be stored in a new variable for the func below
		c.Provisioners[pluginName] = func() (packersdk.Provisioner, error) {
			return c.Plugins.Client(newPath).Provisioner()
		}
		externallyUsed = append(externallyUsed, pluginName)
	}
	if len(externallyUsed) > 0 {
		sort.Strings(externallyUsed)
		log.Printf("using external provisioners %v", externallyUsed)
		externallyUsed = nil
	}

	return nil
}

func (c *config) discoverSingle(glob string) (map[string]string, error) {
	matches, err := filepath.Glob(glob)
	if err != nil {
		return nil, err
	}

	res := make(map[string]string)

	prefix := filepath.Base(glob)
	prefix = prefix[:strings.Index(prefix, "*")]
	for _, match := range matches {
		file := filepath.Base(match)

		// On Windows, ignore any plugins that don't end in .exe.
		// We could do a full PATHEXT parse, but this is probably good enough.
		if runtime.GOOS == "windows" && strings.ToLower(filepath.Ext(file)) != ".exe" {
			log.Printf(
				"[DEBUG] Ignoring plugin match %s, no exe extension",
				match)
			continue
		}

		// If the filename has a ".", trim up to there
		if idx := strings.Index(file, ".exe"); idx >= 0 {
			file = file[:idx]
		}

		// Look for foo-bar-baz. The plugin name is "baz"
		pluginName := file[len(prefix):]
		log.Printf("[DEBUG] Discovered plugin: %s = %s", pluginName, match)
		res[pluginName] = match
	}

	return res, nil
}

func (c *config) discoverInternalComponents() error {
	// Get the packer binary path
	packerPath, err := os.Executable()
	if err != nil {
		log.Printf("[ERR] Error loading exe directory: %s", err)
		return err
	}

	for builder := range command.Builders {
		builder := builder
		_, found := (c.Builders)[builder]
		if !found {
			c.Builders[builder] = func() (packersdk.Builder, error) {
				bin := fmt.Sprintf("%s%splugin%spacker-builder-%s",
					packerPath, PACKERSPACE, PACKERSPACE, builder)
				return c.Plugins.Client(bin).Builder()
			}
		}
	}

	for provisioner := range command.Provisioners {
		provisioner := provisioner
		_, found := (c.Provisioners)[provisioner]
		if !found {
			c.Provisioners[provisioner] = func() (packersdk.Provisioner, error) {
				bin := fmt.Sprintf("%s%splugin%spacker-provisioner-%s",
					packerPath, PACKERSPACE, PACKERSPACE, provisioner)
				return c.Plugins.Client(bin).Provisioner()
			}
		}
	}

	for postProcessor := range command.PostProcessors {
		postProcessor := postProcessor
		_, found := (c.PostProcessors)[postProcessor]
		if !found {
			c.PostProcessors[postProcessor] = func() (packersdk.PostProcessor, error) {
				bin := fmt.Sprintf("%s%splugin%spacker-post-processor-%s",
					packerPath, PACKERSPACE, PACKERSPACE, postProcessor)
				return c.Plugins.Client(bin).PostProcessor()
			}
		}
	}

	return nil
}
