package packer

import pluginsdk "github.com/hashicorp/packer-plugin-sdk/plugin"

var (
	PluginFileExtension = "_x" + pluginsdk.APIVersion + ".exe" // OS-Specific plugin file extention
)
