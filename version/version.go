// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package version

import (
	_ "embed"
	"fmt"
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

	//go:embed VERSION
	rawVersion string

	// The next version number that will be released. This will be updated after every release
	// Version must conform to the format expected by github.com/hashicorp/go-version
	// for tests to work.
	// A pre-release marker for the version can also be specified (e.g -dev). If this is omitted
	// The main version number that is being run at the moment.
	Version string
	// A pre-release marker for the version. If this is "" (empty string)
	// then it means that it is a final release. Otherwise, this is a pre-release
	// such as "dev" (in development), "beta", "rc1", etc.
	VersionPrerelease string
	// VersionMetadata may be added to give more non-normalised information on a build
	// like a commit SHA for example.
	//
	// Ex: 1.0.0-dev+metadata
	VersionMetadata string
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
	var err error

	// Note: we use strings.TrimSpace on the version read from version/VERSION
	// as it could have trailing whitespaces that must not be part of the
	// version string, otherwise version.NewSemver will reject it.
	SemVer, err = version.NewSemver(strings.TrimSpace(rawVersion))
	if err != nil {
		panic(fmt.Sprintf("Invalid semver version specified in 'version/VERSION' (%q): %s", rawVersion, err))
	}

	Version = SemVer.Core().String()
	VersionPrerelease = SemVer.Prerelease()
	VersionMetadata = SemVer.Metadata()

	PackerVersion = pluginVersion.InitializePluginVersion(SemVer.Core().String(), SemVer.Prerelease())
}

// String returns the complete version string, including prerelease
func String() string {
	return PackerVersion.String()
}
