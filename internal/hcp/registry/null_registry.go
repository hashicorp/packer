// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package registry

import (
	"context"

	sdkpacker "github.com/hashicorp/packer-plugin-sdk/packer"
)

// nullRegistry is a special handler that does nothing
type nullRegistry struct{}

func (r nullRegistry) PopulateVersion(context.Context) error {
	return nil
}

func (r nullRegistry) StartBuild(context.Context, sdkpacker.Build) error {
	return nil
}

func (r nullRegistry) CompleteBuild(
	ctx context.Context,
	build sdkpacker.Build,
	artifacts []sdkpacker.Artifact,
	buildErr error,
) ([]sdkpacker.Artifact, error) {
	return artifacts, nil
}

func (r nullRegistry) VersionStatusSummary() {}
