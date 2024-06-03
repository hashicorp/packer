// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package version

import (
	"github.com/hashicorp/packer-plugin-sdk/version"
	packerVersion "github.com/hashicorp/packer/version"
)

var FileProvisionerVersion *version.PluginVersion

func init() {
	FileProvisionerVersion = version.NewPluginVersion(
		packerVersion.Version, packerVersion.VersionPrerelease, packerVersion.VersionMetadata)
}
