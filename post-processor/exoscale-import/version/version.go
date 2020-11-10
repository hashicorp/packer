package version

import (
	"github.com/hashicorp/packer/helper/version"
	packerVersion "github.com/hashicorp/packer/version"
)

var ExoscaleImportPluginVersion = version.PluginVersion{
	Version:           packerVersion.Version,
	VersionPrerelease: packerVersion.VersionPrerelease,
}
