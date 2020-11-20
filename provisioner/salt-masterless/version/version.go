package version

import (
	"github.com/hashicorp/packer/packer-plugin-sdk/version"
	packerVersion "github.com/hashicorp/packer/version"
)

var SaltPluginVersion *version.PluginVersion

func init() {
	SaltPluginVersion = version.InitializePluginVersion(
		packerVersion.Version, packerVersion.VersionPrerelease)
}
