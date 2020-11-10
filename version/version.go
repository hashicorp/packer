package version

import (
	"github.com/hashicorp/go-version"
	pluginVersion "github.com/hashicorp/packer/helper/version"
)

// The git commit that was compiled. This will be filled in by the compiler.
var GitCommit string

// The main version number that is being run at the moment.
const Version = "1.6.6"

// A pre-release marker for the version. If this is "" (empty string)
// then it means that it is a final release. Otherwise, this is a pre-release
// such as "dev" (in development), "beta", "rc1", etc.
const VersionPrerelease = "dev"

var PackerVersion = pluginVersion.PluginVersion{
	Version:           Version,
	VersionPrerelease: VersionPrerelease,
}

func FormattedVersion() string {
	return PackerVersion.FormattedVersion()
}

// SemVer is an instance of version.Version. This has the secondary
// benefit of verifying during tests and init time that our version is a
// proper semantic version, which should always be the case.
var SemVer *version.Version

func init() {
	SemVer = version.Must(version.NewVersion(Version))
}

// String returns the complete version string, including prerelease
func String() string {
	return PackerVersion.String()
}
