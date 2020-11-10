// Version helps plugin creators set and track the plugin version using the same
// convenience functions used by the Packer core.
package version

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/go-version"
)

// The git commit that was compiled. This will be filled in by the compiler.
var GitCommit string

type PluginVersion struct {
	// The main version number that is being run at the moment.
	Version string
	// A pre-release marker for the version. If this is "" (empty string)
	// then it means that it is a final release. Otherwise, this is a pre-release
	// such as "dev" (in development), "beta", "rc1", etc.
	VersionPrerelease string
}

func (p *PluginVersion) FormattedVersion() string {
	var versionString bytes.Buffer
	fmt.Fprintf(&versionString, "%s", p.Version)
	if p.VersionPrerelease != "" {
		fmt.Fprintf(&versionString, "-%s", p.VersionPrerelease)

		if GitCommit != "" {
			fmt.Fprintf(&versionString, " (%s)", GitCommit)
		}
	}

	return versionString.String()
}

func (p *PluginVersion) Semver() *version.Version {
	// SemVer is an instance of version.Version. This has the secondary
	// benefit of verifying during tests and init time that our version is a
	// proper semantic version, which should always be the case.
	SemVer := version.Must(version.NewVersion(p.Version))
	return SemVer
}

// String returns the complete version string, including prerelease
func (p *PluginVersion) String() string {
	if p.VersionPrerelease != "" {
		return fmt.Sprintf("%s-%s", p.Version, p.VersionPrerelease)
	}
	return p.Version
}
