package packer

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/pathing"
	pluginsdk "github.com/hashicorp/packer-plugin-sdk/plugin"
)

// PluginConfig helps load and use packer plugins
type PluginConfig struct {
	KnownPluginFolders []string
	PluginMinPort      int
	PluginMaxPort      int
	Builders           BuilderSet
	Provisioners       ProvisionerSet
	PostProcessors     PostProcessorSet
	DataSources        DatasourceSet
}

// PACKERSPACE is used to represent the spaces that separate args for a command
// without being confused with spaces in the path to the command itself.
const PACKERSPACE = "-PACKERSPACE-"

// Discover discovers plugins.
//
// Search the directory of the executable, then the plugins directory, and
// finally the CWD, in that order. Any conflicts will overwrite previously
// found plugins, in that order.
// Hence, the priority order is the reverse of the search order - i.e., the
// CWD has the highest priority.
func (c *PluginConfig) Discover() error {
	if c.Builders == nil {
		c.Builders = MapOfBuilder{}
	}
	if c.Provisioners == nil {
		c.Provisioners = MapOfProvisioner{}
	}
	if c.PostProcessors == nil {
		c.PostProcessors = MapOfPostProcessor{}
	}
	if c.DataSources == nil {
		c.DataSources = MapOfDatasource{}
	}

	// If we are already inside a plugin process we should not need to
	// discover anything.
	if os.Getenv(pluginsdk.MagicCookieKey) == pluginsdk.MagicCookieValue {
		return nil
	}

	// TODO: use KnownPluginFolders here. TODO probably after JSON is deprecated
	// so that we can keep the current behavior just the way it is.

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
	dir, err := pathing.ConfigDir()
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

func (c *PluginConfig) discoverExternalComponents(path string) error {
	var err error
	log.Printf("[TRACE] discovering plugins in %s", path)

	if !filepath.IsAbs(path) {
		path, err = filepath.Abs(path)
		if err != nil {
			return err
		}
	}
	var externallyUsed []string

	pluginPaths, err := c.discoverSingle(filepath.Join(path, "packer-builder-*"))
	if err != nil {
		return err
	}
	for pluginName, pluginPath := range pluginPaths {
		newPath := pluginPath // this needs to be stored in a new variable for the func below
		c.Builders.Set(pluginName, func() (packersdk.Builder, error) {
			return c.Client(newPath).Builder()
		})
		externallyUsed = append(externallyUsed, pluginName)
	}
	if len(externallyUsed) > 0 {
		sort.Strings(externallyUsed)
		log.Printf("[INFO] using external builders: %v", externallyUsed)
		externallyUsed = nil
	}

	pluginPaths, err = c.discoverSingle(filepath.Join(path, "packer-post-processor-*"))
	if err != nil {
		return err
	}
	for pluginName, pluginPath := range pluginPaths {
		newPath := pluginPath // this needs to be stored in a new variable for the func below
		c.PostProcessors.Set(pluginName, func() (packersdk.PostProcessor, error) {
			return c.Client(newPath).PostProcessor()
		})
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
		c.Provisioners.Set(pluginName, func() (packersdk.Provisioner, error) {
			return c.Client(newPath).Provisioner()
		})
		externallyUsed = append(externallyUsed, pluginName)
	}
	if len(externallyUsed) > 0 {
		sort.Strings(externallyUsed)
		log.Printf("using external provisioners %v", externallyUsed)
		externallyUsed = nil
	}

	pluginPaths, err = c.discoverSingle(filepath.Join(path, "packer-datasource-*"))
	if err != nil {
		return err
	}
	for pluginName, pluginPath := range pluginPaths {
		newPath := pluginPath // this needs to be stored in a new variable for the func below
		c.DataSources.Set(pluginName, func() (packersdk.Datasource, error) {
			return c.Client(newPath).Datasource()
		})
		externallyUsed = append(externallyUsed, pluginName)
	}
	if len(externallyUsed) > 0 {
		sort.Strings(externallyUsed)
		log.Printf("using external datasource %v", externallyUsed)
	}

	pluginPaths, err = c.discoverSingle(filepath.Join(path, "packer-plugin-*"))
	if err != nil {
		return err
	}

	for pluginName, pluginPath := range pluginPaths {
		if err := c.DiscoverMultiPlugin(pluginName, pluginPath); err != nil {
			return err
		}
	}

	return nil
}

func (c *PluginConfig) discoverSingle(glob string) (map[string]string, error) {
	matches, err := filepath.Glob(glob)
	if err != nil {
		return nil, err
	}

	res := make(map[string]string)

	prefix := filepath.Base(glob)
	prefix = prefix[:strings.Index(prefix, "*")]
	for _, match := range matches {
		file := filepath.Base(match)

		// skip folders like packer-plugin-sdk
		if stat, err := os.Stat(file); err == nil && stat.IsDir() {
			continue
		}

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

// DiscoverMultiPlugin takes the description from a multi-component plugin
// binary and makes the plugins available to use in Packer. Each plugin found in the
// binary will be addressable using `${pluginName}-${builderName}` for example.
// pluginName could be manually set. It usually is a cloud name like amazon.
// pluginName can be extrapolated from the filename of the binary; so
// if the "packer-plugin-amazon" binary had an "ebs" builder one could use
// the "amazon-ebs" builder.
func (c *PluginConfig) DiscoverMultiPlugin(pluginName, pluginPath string) error {
	out, err := exec.Command(pluginPath, "describe").Output()
	if err != nil {
		return err
	}
	var desc pluginsdk.SetDescription
	if err := json.Unmarshal(out, &desc); err != nil {
		return err
	}

	pluginPrefix := pluginName + "-"

	for _, builderName := range desc.Builders {
		builderName := builderName // copy to avoid pointer overwrite issue
		key := pluginPrefix + builderName
		if builderName == pluginsdk.DEFAULT_NAME {
			key = pluginName
		}
		c.Builders.Set(key, func() (packersdk.Builder, error) {
			return c.Client(pluginPath, "start", "builder", builderName).Builder()
		})
	}

	if len(desc.Builders) > 0 {
		log.Printf("[INFO] found external %v builders from %s plugin", desc.Builders, pluginName)
	}

	for _, postProcessorName := range desc.PostProcessors {
		postProcessorName := postProcessorName // copy to avoid pointer overwrite issue
		key := pluginPrefix + postProcessorName
		if postProcessorName == pluginsdk.DEFAULT_NAME {
			key = pluginName
		}
		c.PostProcessors.Set(key, func() (packersdk.PostProcessor, error) {
			return c.Client(pluginPath, "start", "post-processor", postProcessorName).PostProcessor()
		})
	}

	if len(desc.PostProcessors) > 0 {
		log.Printf("[INFO] found external %v post-processors from %s plugin", desc.PostProcessors, pluginName)
	}

	for _, provisionerName := range desc.Provisioners {
		provisionerName := provisionerName // copy to avoid pointer overwrite issue
		key := pluginPrefix + provisionerName
		if provisionerName == pluginsdk.DEFAULT_NAME {
			key = pluginName
		}
		c.Provisioners.Set(key, func() (packersdk.Provisioner, error) {
			return c.Client(pluginPath, "start", "provisioner", provisionerName).Provisioner()
		})
	}
	if len(desc.Provisioners) > 0 {
		log.Printf("found external %v provisioner from %s plugin", desc.Provisioners, pluginName)
	}

	for _, datasourceName := range desc.Datasources {
		datasourceName := datasourceName // copy to avoid pointer overwrite issue
		c.DataSources.Set(pluginPrefix+datasourceName, func() (packersdk.Datasource, error) {
			return c.Client(pluginPath, "start", "datasource", datasourceName).Datasource()
		})
	}
	if len(desc.Datasources) > 0 {
		log.Printf("found external %v datasource from %s plugin", desc.Datasources, pluginName)
	}

	return nil
}

func (c *PluginConfig) Client(path string, args ...string) *PluginClient {
	originalPath := path

	// Check for special case using `packer plugin PLUGIN`
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
		log.Printf("[INFO] exec.LookPath: %s : %v. Checking same directory as executable.", path, err)
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

	if strings.Contains(originalPath, PACKERSPACE) {
		log.Printf("[TRACE] Starting internal plugin %s", args[len(args)-1])
	} else {
		log.Printf("[TRACE] Starting external plugin %s %s", path, strings.Join(args, " "))
	}
	var config PluginClientConfig
	config.Cmd = exec.Command(path, args...)
	config.Managed = true
	config.MinPort = c.PluginMinPort
	config.MaxPort = c.PluginMaxPort
	return NewClient(&config)
}
