// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

package registry

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	sdkpacker "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer"
)

// nullRegistry is a special handler that does nothing
type nullRegistry struct{}

func (r nullRegistry) PopulateVersion(context.Context) error {
	return nil
}

func (r nullRegistry) StartBuild(context.Context, *packer.CoreBuild) error {
	return nil
}

func (r nullRegistry) CompleteBuild(
	ctx context.Context,
	build *packer.CoreBuild,
	artifacts []sdkpacker.Artifact,
	buildErr error,
) ([]sdkpacker.Artifact, error) {
	return artifacts, nil
}

func (r nullRegistry) VersionStatusSummary() {}

func (r nullRegistry) Metadata() Metadata {
	return NilMetadata{}
}

func (r nullRegistry) FetchEnforcedBlocks(ctx context.Context) error {
	return nil
}

func (r nullRegistry) InjectEnforcedProvisioners(builds []*packer.CoreBuild) hcl.Diagnostics {
	return nil
}
