package version

import (
	"github.com/hashicorp/packer/packer-plugin-sdk/version"
	packerVersion "github.com/hashicorp/packer/version"
)

var AWSPluginVersion *version.PluginVersion

func init() {
	AWSPluginVersion = version.InitializePluginVersion(
		packerVersion.Version, packerVersion.VersionPrerelease)
}
