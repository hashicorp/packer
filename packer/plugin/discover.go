package plugin

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/hashicorp/packer/packer"
)

// PACKERSPACE is used to represent the spaces that separate args for a command
// without being confused with spaces in the path to the command itself.
const PACKERSPACE = "-PACKERSPACE-"

type Config struct {
	PluginMinPort  int
	PluginMaxPort  int
	Builders       packer.MapOfBuilder       `json:"-"`
	Provisioners   packer.MapOfProvisioner   `json:"-"`
	PostProcessors packer.MapOfPostProcessor `json:"-"`
}

func (c *Config) GetPlugins() (packer.MapOfBuilder, packer.MapOfProvisioner, packer.MapOfPostProcessor) {
	return c.Builders, c.Provisioners, c.PostProcessors
}

// Discover discovers plugins.
//
// Search the directory of the executable, then the plugins directory, and
// finally the CWD, in that order. Any conflicts will overwrite previously
// found plugins, in that order.
// Hence, the priority order is the reverse of the search order - i.e., the
// CWD has the highest priority.
func (c *Config) Discover() error {
	c.Builders = packer.MapOfBuilder{}
	c.PostProcessors = packer.MapOfPostProcessor{}
	c.Provisioners = packer.MapOfProvisioner{}
	// If we are already inside a plugin process we should not need to
	// discover anything.
	if os.Getenv(MagicCookieKey) == MagicCookieValue {
		return nil
	}

	// Next, look in the same directory as the executable.
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("[ERR] Error loading exe directory: %s", err)
	} else {
		if err := c.discoverExternalComponents(filepath.Dir(exePath)); err != nil {
			return err
		}
	}

	// Next, look in the default plugins directory inside the configdir/.packer.d/plugins.
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

	// Check whether there is a custom Plugin directory defined. This gets
	// absolute preference.
	if packerPluginPath := os.Getenv("PACKER_PLUGIN_PATH"); packerPluginPath != "" {
		sep := ":"
		if runtime.GOOS == "windows" {
			// on windows, PATH is semicolon-separated
			sep = ";"
		}
		plugPaths := strings.Split(packerPluginPath, sep)
		for _, plugPath := range plugPaths {
			if err := c.discoverExternalComponents(plugPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Config) discoverExternalComponents(path string) error {
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
			return c.Client(newPath).Builder()
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
			return c.Client(newPath).PostProcessor()
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
			return c.Client(newPath).Provisioner()
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

func (c *Config) discoverSingle(glob string) (map[string]string, error) {
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

func (c *Config) Client(path string) *Client {
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
		exePath, err := os.Executable()
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
	var config ClientConfig
	config.Cmd = exec.Command(path, args...)
	config.Managed = true
	config.MinPort = c.PluginMinPort
	config.MaxPort = c.PluginMaxPort
	return NewClient(&config)
}
