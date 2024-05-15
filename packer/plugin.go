// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package packer

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	pluginsdk "github.com/hashicorp/packer-plugin-sdk/plugin"
	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
)

var defaultChecksummer = plugingetter.Checksummer{
	Type: "sha256",
	Hash: sha256.New(),
}

// PluginConfig helps load and use packer plugins
type PluginConfig struct {
	PluginDirectory string
	PluginMinPort   int
	PluginMaxPort   int
	Builders        BuilderSet
	Provisioners    ProvisionerSet
	PostProcessors  PostProcessorSet
	DataSources     DatasourceSet
	ReleasesOnly    bool
}

// PACKERSPACE is used to represent the spaces that separate args for a command
// without being confused with spaces in the path to the command itself.
const PACKERSPACE = "-PACKERSPACE-"

var extractPluginBasename = regexp.MustCompile("^packer-plugin-([^_]+)")

// Discover discovers the latest installed version of each installed plugin.
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

	if c.PluginDirectory == "" {
		c.PluginDirectory, _ = PluginFolder()
	}

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	installations, err := plugingetter.Requirement{}.ListInstallations(plugingetter.ListInstallationsOptions{
		PluginDirectory: c.PluginDirectory,
		BinaryInstallationOptions: plugingetter.BinaryInstallationOptions{
			OS:              runtime.GOOS,
			ARCH:            runtime.GOARCH,
			Ext:             ext,
			APIVersionMajor: pluginsdk.APIVersionMajor,
			APIVersionMinor: pluginsdk.APIVersionMinor,
			Checksummers: []plugingetter.Checksummer{
				{Type: "sha256", Hash: sha256.New()},
			},
			ReleasesOnly: c.ReleasesOnly,
		},
	})
	if err != nil {
		return err
	}

	// Map of plugin basename to executable
	//
	// We'll use that later to register the components for each plugin
	pluginMap := map[string]string{}
	for _, install := range installations {
		pluginBasename := filepath.Base(install.BinaryPath)
		matches := extractPluginBasename.FindStringSubmatch(pluginBasename)
		if len(matches) != 2 {
			log.Printf("[INFO] - plugin %q could not have its name matched, ignoring", pluginBasename)
			continue
		}

		pluginName := matches[1]
		pluginMap[pluginName] = install.BinaryPath
	}

	for name, path := range pluginMap {
		err := c.DiscoverMultiPlugin(name, path)
		if err != nil {
			return err
		}
	}

	return nil
}

// DiscoverMultiPlugin takes the description from a multi-component plugin
// binary and makes the plugins available to use in Packer. Each plugin found in the
// binary will be addressable using `${pluginName}-${builderName}` for example.
// pluginName could be manually set. It usually is a cloud name like amazon.
// pluginName can be extrapolated from the filename of the binary; so
// if the "packer-plugin-amazon" binary had an "ebs" builder one could use
// the "amazon-ebs" builder.
func (c *PluginConfig) DiscoverMultiPlugin(pluginName, pluginPath string) error {
	desc, err := plugingetter.GetPluginDescription(pluginPath)
	if err != nil {
		return fmt.Errorf("failed to get plugin description from executable %q: %s", pluginPath, err)
	}

	pluginPrefix := pluginName + "-"
	pluginDetails := PluginDetails{
		Name:        pluginName,
		Description: desc,
		PluginPath:  pluginPath,
	}

	for _, builderName := range desc.Builders {
		builderName := builderName // copy to avoid pointer overwrite issue
		key := pluginPrefix + builderName
		if builderName == pluginsdk.DEFAULT_NAME {
			key = pluginName
		}
		if c.Builders.Has(key) {
			continue
		}
		c.Builders.Set(key, func() (packersdk.Builder, error) {
			return c.Client(pluginPath, "start", "builder", builderName).Builder()
		})
		GlobalPluginsDetailsStore.SetBuilder(key, pluginDetails)
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
		if c.PostProcessors.Has(key) {
			continue
		}
		c.PostProcessors.Set(key, func() (packersdk.PostProcessor, error) {
			return c.Client(pluginPath, "start", "post-processor", postProcessorName).PostProcessor()
		})
		GlobalPluginsDetailsStore.SetPostProcessor(key, pluginDetails)
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
		if c.Provisioners.Has(key) {
			continue
		}
		c.Provisioners.Set(key, func() (packersdk.Provisioner, error) {
			return c.Client(pluginPath, "start", "provisioner", provisionerName).Provisioner()
		})
		GlobalPluginsDetailsStore.SetProvisioner(key, pluginDetails)

	}
	if len(desc.Provisioners) > 0 {
		log.Printf("found external %v provisioner from %s plugin", desc.Provisioners, pluginName)
	}

	for _, datasourceName := range desc.Datasources {
		datasourceName := datasourceName // copy to avoid pointer overwrite issue
		key := pluginPrefix + datasourceName
		if datasourceName == pluginsdk.DEFAULT_NAME {
			key = pluginName
		}
		if c.DataSources.Has(key) {
			continue
		}
		c.DataSources.Set(key, func() (packersdk.Datasource, error) {
			return c.Client(pluginPath, "start", "datasource", datasourceName).Datasource()
		})
		GlobalPluginsDetailsStore.SetDataSource(key, pluginDetails)
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
		log.Printf("[INFO] Starting internal plugin %s", args[len(args)-1])
	} else {
		log.Printf("[INFO] Starting external plugin %s %s", path, strings.Join(args, " "))
	}
	var config PluginClientConfig
	config.Cmd = exec.Command(path, args...)
	config.Managed = true
	config.MinPort = c.PluginMinPort
	config.MaxPort = c.PluginMaxPort
	return NewClient(&config)
}

type PluginComponentType string

const (
	PluginComponentBuilder       PluginComponentType = "builder"
	PluginComponentPostProcessor PluginComponentType = "post-processor"
	PluginComponentProvisioner   PluginComponentType = "provisioner"
	PluginComponentDataSource    PluginComponentType = "data-source"
)

type PluginDetails struct {
	Name        string
	Description pluginsdk.SetDescription
	PluginPath  string
}

type pluginsDetailsStorage struct {
	rwMutex sync.RWMutex
	data    map[string]PluginDetails
}

var GlobalPluginsDetailsStore = &pluginsDetailsStorage{
	data: make(map[string]PluginDetails),
}

func (pds *pluginsDetailsStorage) set(key string, plugin PluginDetails) {
	pds.rwMutex.Lock()
	defer pds.rwMutex.Unlock()
	pds.data[key] = plugin
}

func (pds *pluginsDetailsStorage) get(key string) (PluginDetails, bool) {
	pds.rwMutex.RLock()
	defer pds.rwMutex.RUnlock()
	plugin, exists := pds.data[key]
	return plugin, exists
}

func (pds *pluginsDetailsStorage) SetBuilder(name string, plugin PluginDetails) {
	key := fmt.Sprintf("%q-%q", PluginComponentBuilder, name)
	pds.set(key, plugin)
}

func (pds *pluginsDetailsStorage) GetBuilder(name string) (PluginDetails, bool) {
	key := fmt.Sprintf("%q-%q", PluginComponentBuilder, name)
	return pds.get(key)
}

func (pds *pluginsDetailsStorage) SetPostProcessor(name string, plugin PluginDetails) {
	key := fmt.Sprintf("%q-%q", PluginComponentPostProcessor, name)
	pds.set(key, plugin)
}

func (pds *pluginsDetailsStorage) GetPostProcessor(name string) (PluginDetails, bool) {
	key := fmt.Sprintf("%q-%q", PluginComponentPostProcessor, name)
	return pds.get(key)
}

func (pds *pluginsDetailsStorage) SetProvisioner(name string, plugin PluginDetails) {
	key := fmt.Sprintf("%q-%q", PluginComponentProvisioner, name)
	pds.set(key, plugin)
}

func (pds *pluginsDetailsStorage) GetProvisioner(name string) (PluginDetails, bool) {
	key := fmt.Sprintf("%q-%q", PluginComponentProvisioner, name)
	return pds.get(key)
}

func (pds *pluginsDetailsStorage) SetDataSource(name string, plugin PluginDetails) {
	key := fmt.Sprintf("%q-%q", PluginComponentDataSource, name)
	pds.set(key, plugin)
}

func (pds *pluginsDetailsStorage) GetDataSource(name string) (PluginDetails, bool) {
	key := fmt.Sprintf("%q-%q", PluginComponentDataSource, name)
	return pds.get(key)
}
