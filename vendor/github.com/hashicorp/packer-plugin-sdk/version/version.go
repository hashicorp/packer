// Package version helps plugin creators set and track the plugin version using
// the same convenience functions used by the Packer core.
package version

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/go-version"
)

// The git commit that was compiled. This will be filled in by the compiler.
var GitCommit string

// Package version helps plugin creators set and track the sdk version using
var Version = "0.0.14"

// A pre-release marker for the version. If this is "" (empty string)
// then it means that it is a final release. Otherwise, this is a pre-release
// such as "dev" (in development), "beta", "rc1", etc.
var VersionPrerelease = ""

// SDKVersion is used by the plugin set to allow Packer to recognize
// what version of the sdk the plugin is.
var SDKVersion = InitializePluginVersion(Version, VersionPrerelease)

// InitializePluginVersion initializes the SemVer and returns a version var.
// If the provided "version" string is not valid, the call to version.Must
// will panic. Therefore, this function should always be called in a package
// init() function to make sure that plugins are following proper semantic
// versioning and to make sure that plugins which aren't following proper
// semantic versioning crash immediately rather than later.
func InitializePluginVersion(vers, versionPrerelease string) *PluginVersion {
	if vers == "" {
		// Defaults to "0.0.0". Useful when binary is created for development purpose.
		vers = "0.0.0"
	}
	pv := PluginVersion{
		version:           vers,
		versionPrerelease: versionPrerelease,
	}
	// This call initializes the SemVer to make sure that if Packer crashes due
	// to an invalid SemVer it's at the very beginning of the Packer run.
	pv.semVer = version.Must(version.NewVersion(vers))
	return &pv
}

type PluginVersion struct {
	// The main version number that is being run at the moment.
	version string
	// A pre-release marker for the version. If this is "" (empty string)
	// then it means that it is a final release. Otherwise, this is a pre-release
	// such as "dev" (in development), "beta", "rc1", etc.
	versionPrerelease string
	// The Semantic Version of the plugin. Used for version constraint comparisons
	semVer *version.Version
}

func (p *PluginVersion) FormattedVersion() string {
	var versionString bytes.Buffer
	fmt.Fprintf(&versionString, "%s", p.version)
	if p.versionPrerelease != "" {
		fmt.Fprintf(&versionString, "-%s", p.versionPrerelease)

		if GitCommit != "" {
			fmt.Fprintf(&versionString, " (%s)", GitCommit)
		}
	}

	return versionString.String()
}

func (p *PluginVersion) SemVer() *version.Version {
	if p.semVer != nil {
		// SemVer is an instance of version.Version. This has the secondary
		// benefit of verifying during tests and init time that our version is a
		// proper semantic version, which should always be the case.
		p.semVer = version.Must(version.NewVersion(p.version))
	}
	return p.semVer
}

func (p *PluginVersion) GetVersion() string {
	return p.version
}

func (p *PluginVersion) GetVersionPrerelease() string {
	return p.versionPrerelease
}

// String returns the complete version string, including prerelease
func (p *PluginVersion) String() string {
	if p.versionPrerelease != "" {
		return fmt.Sprintf("%s-%s", p.version, p.versionPrerelease)
	}
	return p.version
}
