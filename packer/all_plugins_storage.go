package packer

import (
	"fmt"

	pluginsdk "github.com/hashicorp/packer-plugin-sdk/plugin"
)

type PluginComponentType string

const (
	PluginComponentBuilder       PluginComponentType = "builder"
	PluginComponentPostProcessor PluginComponentType = "post-processor"
	PluginComponentProvisioner   PluginComponentType = "provisioner"
)

type PluginDetails struct {
	Name        string
	Description pluginsdk.SetDescription
	PluginPath  string
}

type allPluginsStorage struct {
	Components map[string]*PluginDetails
}

var (
	AllPluginsStorage *allPluginsStorage
)

func init() {
	AllPluginsStorage = &allPluginsStorage{
		Components: map[string]*PluginDetails{},
	}
}

func (aps *allPluginsStorage) AddPluginDetails(
	componentType PluginComponentType, componentKey, pluginName, pluginPath string, pluginDescription pluginsdk.SetDescription,
) {
	key := fmt.Sprintf("%q-%q", componentKey, componentType)
	aps.Components[key] = &PluginDetails{
		Name:        pluginName,
		Description: pluginDescription,
		PluginPath:  pluginPath,
	}
}

func (aps *allPluginsStorage) GetPluginDetailsFor(componentType PluginComponentType, componentKey string) *PluginDetails {
	key := fmt.Sprintf("%q-%q", componentKey, componentType)
	pluginDetails, ok := aps.Components[key]
	if !ok {
		return nil
	}
	return pluginDetails
}
