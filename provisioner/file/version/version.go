// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package version

import (
	"github.com/hashicorp/packer-plugin-sdk/version"
	packerVersion "github.com/hashicorp/packer/version"
)

var FileProvisionerVersion *version.PluginVersion

func init() {
	FileProvisionerVersion = version.InitializePluginVersion(
		packerVersion.Version, packerVersion.VersionPrerelease)
}
