package metadata

import (
	"sync"

	pluginsdk "github.com/hashicorp/packer-plugin-sdk/plugin"
)

type PluginDetails struct {
	Name        string
	Description pluginsdk.SetDescription
}

type AllPluginsStorage struct {
	Components map[string]*PluginDetails
}

var (
	allPluginsStorage     *AllPluginsStorage
	allPluginsStorageOnce sync.Once
)

func GetAllPluginsStorage() *AllPluginsStorage {
	allPluginsStorageOnce.Do(func() {
		allPluginsStorage = &AllPluginsStorage{
			Components: map[string]*PluginDetails{},
		}
	})
	return allPluginsStorage
}

func (aps *AllPluginsStorage) AddPluginDetails(componentKey, pluginName string, pluginDescription pluginsdk.SetDescription) {
	aps.Components[componentKey] = &PluginDetails{
		Name:        pluginName,
		Description: pluginDescription,
	}
}

func (aps *AllPluginsStorage) GetPluginDetailsFor(componentKey string) *PluginDetails {
	pluginDetails, ok := aps.Components[componentKey]
	if !ok {
		return nil
	}
	return pluginDetails
}
