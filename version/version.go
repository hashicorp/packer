package version

import (
	_ "embed"
	"strings"

	"github.com/hashicorp/go-version"
	pluginVersion "github.com/hashicorp/packer-plugin-sdk/version"
)

var (
	// The git commit that was compiled. This will be filled in by the compiler.
	GitCommit   string
	GitDescribe string

	// Whether cgo is enabled or not; set at build time
	CgoEnabled bool

	// The next version number that will be released. This will be updated after every release
	// Version must conform to the format expected by github.com/hashicorp/go-version
	// for tests to work.
	// A pre-release marker for the version can also be specified (e.g -dev). If this is omitted
	// then it means that it is a final release. Otherwise, this is a pre-release
	// such as "dev" (in development), "beta", "rc1", etc.
	//go:embed VERSION
	fullVersion string

	Version, VersionPrerelease, _ = strings.Cut(fullVersion, "-")

	VersionMetadata = ""
)

var PackerVersion *pluginVersion.PluginVersion

func FormattedVersion() string {
	return PackerVersion.FormattedVersion()
}

// SemVer is an instance of version.Version. This has the secondary
// benefit of verifying during tests and init time that our version is a
// proper semantic version, which should always be the case.
var SemVer *version.Version

func init() {
	PackerVersion = pluginVersion.InitializePluginVersion(Version, VersionPrerelease)
	SemVer = PackerVersion.SemVer()
}

// String returns the complete version string, including prerelease
func String() string {
	return PackerVersion.String()
}
