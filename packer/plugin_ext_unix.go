// +build aix darwin dragonfly freebsd js,wasm linux netbsd openbsd solaris

package packer

import pluginsdk "github.com/hashicorp/packer-plugin-sdk/plugin"

var (
	PluginFileExtension = "_x" + pluginsdk.APIVersion // OS-Specific plugin file extention
)
