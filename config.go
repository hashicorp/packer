package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/hashicorp/packer/command"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer/plugin"
	"github.com/kardianos/osext"
)

// PACKERSPACE is used to represent the spaces that separate args for a command
// without being confused with spaces in the path to the command itself.
const PACKERSPACE = "-PACKERSPACE-"

type config struct {
	DisableCheckpoint          bool `json:"disable_checkpoint"`
	DisableCheckpointSignature bool `json:"disable_checkpoint_signature"`
	PluginMinPort              int
	PluginMaxPort              int

	Builders       packer.MapOfBuilder
	Provisioners   packer.MapOfProvisioner
	PostProcessors packer.MapOfPostProcessor `json:"post-processors"`
}

// DecodeConfig decodes configuration in JSON format from the given io.Reader into
// the config object pointed to.
func DecodeConfig(r io.Reader, c *config) error {
	decoder := json.NewDecoder(r)
	return decoder.Decode(c)
}

// Discover discovers plugins.
//
// Search the directory of the executable, then the plugins directory, and
// finally the CWD, in that order. Any conflicts will overwrite previously
// found plugins, in that order.
// Hence, the priority order is the reverse of the search order - i.e., the
// CWD has the highest priority.
func (c *config) Discover() error {
	// If we are already inside a plugin process we should not need to
	// discover anything.
	if os.Getenv(plugin.MagicCookieKey) == plugin.MagicCookieValue {
		return nil
	}

	// First, look in the same directory as the executable.
	exePath, err := osext.Executable()
	if err != nil {
		log.Printf("[ERR] Error loading exe directory: %s", err)
	} else {
		if err := c.discoverExternalComponents(filepath.Dir(exePath)); err != nil {
			return err
		}
	}

	// Next, look in the plugins directory.
	dir, err := packer.ConfigDir()
	if err != nil {
		log.Printf("[ERR] Error loading config directory: %s", err)
	} else {
		if err := c.discoverExternalComponents(filepath.Join(dir, "plugins")); err != nil {
			return err
		}
	}

	// Next, look in the CWD.
	if err := c.discoverExternalComponents("."); err != nil {
		return err
	}

	// Finally, try to use an internal plugin. Note that this will not override
	// any previously-loaded plugins.
	if err := c.discoverInternalComponents(); err != nil {
		return err
	}

	return nil
}

// This is a proper packer.BuilderFunc that can be used to load packer.Builder
// implementations from the defined plugins.
func (c *config) StartBuilder(name string) (packer.Builder, error) {
	log.Printf("Loading builder: %s\n", name)
	return c.Builders.Start(name)
}

// This is a proper implementation of packer.HookFunc that can be used
// to load packer.Hook implementations from the defined plugins.
func (c *config) StarHook(name string) (packer.Hook, error) {
	log.Printf("Loading hook: %s\n", name)
	return c.pluginClient(name).Hook()
}

// This is a proper packer.PostProcessorFunc that can be used to load
// packer.PostProcessor implementations from defined plugins.
func (c *config) StartPostProcessor(name string) (packer.PostProcessor, error) {
	log.Printf("Loading post-processor: %s", name)
	return c.PostProcessors.Start(name)
}

// This is a proper packer.ProvisionerFunc that can be used to load
// packer.Provisioner implementations from defined plugins.
func (c *config) StartProvisioner(name string) (packer.Provisioner, error) {
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
		c.Builders[pluginName] = func() (packer.Builder, error) {
			return c.pluginClient(newPath).Builder()
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
		c.PostProcessors[pluginName] = func() (packer.PostProcessor, error) {
			return c.pluginClient(newPath).PostProcessor()
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
		c.Provisioners[pluginName] = func() (packer.Provisioner, error) {
			return c.pluginClient(newPath).Provisioner()
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

		// One Windows, ignore any plugins that don't end in .exe.
		// We could do a full PATHEXT parse, but this is probably good enough.
		if runtime.GOOS == "windows" && strings.ToLower(filepath.Ext(file)) != ".exe" {
			log.Printf(
				"[DEBUG] Ignoring plugin match %s, no exe extension",
				match)
			continue
		}

		// If the filename has a ".", trim up to there
		if idx := strings.Index(file, "."); idx >= 0 {
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
	packerPath, err := osext.Executable()
	if err != nil {
		log.Printf("[ERR] Error loading exe directory: %s", err)
		return err
	}

	for builder := range command.Builders {
		builder := builder
		_, found := (c.Builders)[builder]
		if !found {
			c.Builders[builder] = func() (packer.Builder, error) {
				bin := fmt.Sprintf("%s%splugin%spacker-builder-%s",
					packerPath, PACKERSPACE, PACKERSPACE, builder)
				return c.pluginClient(bin).Builder()
			}
		}
	}

	for provisioner := range command.Provisioners {
		provisioner := provisioner
		_, found := (c.Provisioners)[provisioner]
		if !found {
			c.Provisioners[provisioner] = func() (packer.Provisioner, error) {
				bin := fmt.Sprintf("%s%splugin%spacker-provisioner-%s",
					packerPath, PACKERSPACE, PACKERSPACE, provisioner)
				return c.pluginClient(bin).Provisioner()
			}
		}
	}

	for postProcessor := range command.PostProcessors {
		postProcessor := postProcessor
		_, found := (c.PostProcessors)[postProcessor]
		if !found {
			c.PostProcessors[postProcessor] = func() (packer.PostProcessor, error) {
				bin := fmt.Sprintf("%s%splugin%spacker-post-processor-%s",
					packerPath, PACKERSPACE, PACKERSPACE, postProcessor)
				return c.pluginClient(bin).PostProcessor()
			}
		}
	}

	return nil
}

func (c *config) pluginClient(path string) *plugin.Client {
	originalPath := path

	// Check for special case using `packer plugin PLUGIN`
	args := []string{}
	if strings.Contains(path, PACKERSPACE) {
		parts := strings.Split(path, PACKERSPACE)
		path = parts[0]
		args = parts[1:]
	}

	// First attempt to find the executable by consulting the PATH.
	path, err := exec.LookPath(path)
	if err != nil {
		// If that doesn't work, look for it in the same directory
		// as the `packer` executable (us).
		log.Printf("Plugin could not be found at %s (%v). Checking same directory as executable.", originalPath, err)
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
	config.Cmd = exec.Command(path, args...)
	config.Managed = true
	config.MinPort = c.PluginMinPort
	config.MaxPort = c.PluginMaxPort
	return plugin.NewClient(&config)
}
