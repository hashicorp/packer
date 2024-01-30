// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package version

import (
	"github.com/hashicorp/packer-plugin-sdk/version"
	packerVersion "github.com/hashicorp/packer/version"
)

var CompressPluginVersion *version.PluginVersion

func init() {
	CompressPluginVersion = version.InitializePluginVersion(
		packerVersion.Version, packerVersion.VersionPrerelease)
}
