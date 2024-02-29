package packer

import (
	pluginsdk "github.com/hashicorp/packer-plugin-sdk/plugin"
)

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

var PluginsDetailsStorage = map[string]PluginDetails{}

func AddPluginDetails(
	componentKey, pluginName, pluginPath string, pluginDescription pluginsdk.SetDescription,
) {
	PluginsDetailsStorage[componentKey] = PluginDetails{
		Name:        pluginName,
		Description: pluginDescription,
		PluginPath:  pluginPath,
	}
}
