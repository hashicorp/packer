package version

import (
	"github.com/hashicorp/packer-plugin-sdk/version"
	packerVersion "github.com/hashicorp/packer/version"
)

var VSphereTemplatePostprocessorVersion *version.PluginVersion

func init() {
	VSphereTemplatePostprocessorVersion = version.InitializePluginVersion(
		packerVersion.Version, packerVersion.VersionPrerelease)
}
