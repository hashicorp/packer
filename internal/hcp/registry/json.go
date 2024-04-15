// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package registry

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	sdkpacker "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer"
)

// JSONRegistry is a HCP handler made to process legacy JSON templates
type JSONRegistry struct {
	configuration *packer.Core
	bucket        *Bucket
	ui            sdkpacker.Ui
}

func NewJSONRegistry(config *packer.Core, ui sdkpacker.Ui) (*JSONRegistry, hcl.Diagnostics) {
	bucket, diags := createConfiguredBucket(
		filepath.Dir(config.Template.Path),
		withPackerEnvConfiguration,
	)

	if diags.HasErrors() {
		return nil, diags
	}

	for _, b := range config.Template.Builders {
		buildName := b.Name

		// By default, if the name is unspecified, it will be assigned the type
		//
		// If the two are different, we can compose the HCP build name from both
		if b.Name != b.Type {
			buildName = fmt.Sprintf("%s.%s", b.Type, b.Name)
		}

		// Get all builds slated within config ignoring any only or exclude flags.
		bucket.RegisterBuildForComponent(buildName)
	}

	ui.Say(fmt.Sprintf("Tracking build on HCP Packer with fingerprint %q", bucket.Version.Fingerprint))

	return &JSONRegistry{
		configuration: config,
		bucket:        bucket,
		ui:            ui,
	}, nil
}

// PopulateVersion creates the metadata in HCP Packer Registry for a build
func (h *JSONRegistry) PopulateVersion(ctx context.Context) error {
	err := h.bucket.Validate()
	if err != nil {
		return err
	}
	err = h.bucket.Initialize(ctx, hcpPackerModels.HashicorpCloudPacker20230101TemplateTypeJSON)
	if err != nil {
		return err
	}

	err = h.bucket.populateVersion(ctx)
	if err != nil {
		return err
	}

	sha, err := getGitSHA(h.configuration.Template.Path)
	if err != nil {
		log.Printf("failed to get GIT SHA from environment, won't set as build labels")
	} else {
		h.bucket.Version.AddSHAToBuildLabels(sha)
	}

	return nil
}

// StartBuild is invoked when one build for the configuration is starting to be processed
func (h *JSONRegistry) StartBuild(ctx context.Context, build sdkpacker.Build) error {
	name := build.Name()
	return h.bucket.startBuild(ctx, name)
}

// CompleteBuild is invoked when one build for the configuration has finished
func (h *JSONRegistry) CompleteBuild(
	ctx context.Context,
	build sdkpacker.Build,
	artifacts []sdkpacker.Artifact,
	buildErr error,
) ([]sdkpacker.Artifact, error) {
	buildName := build.Name()
	buildMetadata := build.(*packer.CoreBuild).GetMetadata()
	err := h.bucket.Version.AddMetadataToBuild(ctx, buildName, buildMetadata)
	if err != nil {
		return nil, err
	}
	return h.bucket.completeBuild(ctx, buildName, artifacts, buildErr)
}

// VersionStatusSummary prints a status report in the UI if the version is not yet done
func (h *JSONRegistry) VersionStatusSummary() {
	h.bucket.Version.statusSummary(h.ui)
}
