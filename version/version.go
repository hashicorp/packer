// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

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
	rawVersion = strings.TrimSpace(rawVersion)

	PackerVersion = pluginVersion.NewRawVersion(rawVersion)
	// A bug in the SDK prevents us from calling SemVer on the PluginVersion
	// derived from the rawVersion, as when doing so, we reset the semVer
	// attribute to only use the core part of the version, thereby dropping any
	// information on pre-release/metadata.
	SemVer, _ = version.NewVersion(rawVersion)

	Version = PackerVersion.GetVersion()
	VersionPrerelease = PackerVersion.GetVersionPrerelease()
	VersionMetadata = PackerVersion.GetMetadata()
}

// String returns the complete version string, including prerelease
func String() string {
	return PackerVersion.String()
}
