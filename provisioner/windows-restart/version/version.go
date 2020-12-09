package version

import (
	"github.com/hashicorp/packer-plugin-sdk/version"
	packerVersion "github.com/hashicorp/packer/version"
)

var WindowsRestartPluginVersion *version.PluginVersion

func init() {
	WindowsRestartPluginVersion = version.InitializePluginVersion(
		packerVersion.Version, packerVersion.VersionPrerelease)
}
