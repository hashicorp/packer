// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package version

import (
	"github.com/hashicorp/packer-plugin-sdk/version"
	packerVersion "github.com/hashicorp/packer/version"
)

var WindowsShellPluginVersion *version.PluginVersion

func init() {
	WindowsShellPluginVersion = version.NewPluginVersion(
		packerVersion.Version, packerVersion.VersionPrerelease, packerVersion.VersionMetadata)
}
