package packer

import (
	pluginsdk "github.com/hashicorp/packer-plugin-sdk/plugin"
)

type PluginDetails struct {
	Name        string
	Description pluginsdk.SetDescription
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

func (aps *allPluginsStorage) AddPluginDetails(componentKey, pluginName string, pluginDescription pluginsdk.SetDescription) {
	aps.Components[componentKey] = &PluginDetails{
		Name:        pluginName,
		Description: pluginDescription,
	}
}

func (aps *allPluginsStorage) GetPluginDetailsFor(componentKey string) *PluginDetails {
	pluginDetails, ok := aps.Components[componentKey]
	if !ok {
		return nil
	}
	return pluginDetails
}
