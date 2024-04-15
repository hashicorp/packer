// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package registry

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2"
	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	sdkpacker "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/hcl2template"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
)

// HCLRegistry is a HCP handler made for handling HCL configurations
type HCLRegistry struct {
	configuration *hcl2template.PackerConfig
	bucket        *Bucket
	ui            sdkpacker.Ui
}

const (
	// Known HCP Packer Datasource, whose id is the SourceImageId for some build.
	hcpImageDatasourceType    string = "hcp-packer-image"
	hcpArtifactDatasourceType string = "hcp-packer-artifact"

	hcpIterationDatasourceType string = "hcp-packer-iteration"
	hcpVersionDatasourceType   string = "hcp-packer-version"

	buildLabel string = "build"
)

// PopulateVersion creates the metadata in HCP Packer Registry for a build
func (h *HCLRegistry) PopulateVersion(ctx context.Context) error {
	err := h.bucket.Initialize(ctx, hcpPackerModels.HashicorpCloudPacker20230101TemplateTypeHCL2)
	if err != nil {
		return err
	}

	err = h.bucket.populateVersion(ctx)
	if err != nil {
		return err
	}

	versionID := h.bucket.Version.ID
	versionFingerprint := h.bucket.Version.Fingerprint

	// FIXME: Remove
	h.configuration.HCPVars["iterationID"] = cty.StringVal(versionID)
	h.configuration.HCPVars["versionFingerprint"] = cty.StringVal(versionFingerprint)

	sha, err := getGitSHA(h.configuration.Basedir)
	if err != nil {
		log.Printf("failed to get GIT SHA from environment, won't set as build labels")
	} else {
		h.bucket.Version.AddSHAToBuildLabels(sha)
	}

	return nil
}

// StartBuild is invoked when one build for the configuration is starting to be processed
func (h *HCLRegistry) StartBuild(ctx context.Context, build sdkpacker.Build) error {
	name := build.Name()
	cb, ok := build.(*packer.CoreBuild)
	if ok {
		name = cb.Type
	}

	return h.bucket.startBuild(ctx, name)
}

// CompleteBuild is invoked when one build for the configuration has finished
func (h *HCLRegistry) CompleteBuild(
	ctx context.Context,
	build sdkpacker.Build,
	artifacts []sdkpacker.Artifact,
	buildErr error,
) ([]sdkpacker.Artifact, error) {
	buildName := build.Name()
	cb, ok := build.(*packer.CoreBuild)
	if ok {
		buildName = cb.Type
	}

	metadata := cb.GetMetadata()
	err := h.bucket.Version.AddMetadataToBuild(ctx, buildName, metadata)
	if err != nil {
		return nil, err
	}
	return h.bucket.completeBuild(ctx, buildName, artifacts, buildErr)
}

// VersionStatusSummary prints a status report in the UI if the version is not yet done
func (h *HCLRegistry) VersionStatusSummary() {
	h.bucket.Version.statusSummary(h.ui)
}

func NewHCLRegistry(config *hcl2template.PackerConfig, ui sdkpacker.Ui) (*HCLRegistry, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	if len(config.Builds) > 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Multiple " + buildLabel + " blocks",
			Detail: fmt.Sprintf("For HCP Packer Registry enabled builds, only one " + buildLabel +
				" block can be defined. Please remove any additional " + buildLabel +
				" block(s). If this " + buildLabel + " is not meant for the HCP Packer registry please " +
				"clear any HCP_PACKER_* environment variables."),
		})

		return nil, diags
	}

	withHCLBucketConfiguration := func(bb *hcl2template.BuildBlock) bucketConfigurationOpts {
		return func(bucket *Bucket) hcl.Diagnostics {
			bucket.ReadFromHCLBuildBlock(bb)
			// If at this point the bucket.Name is still empty,
			// last try is to use the build.Name if present
			if bucket.Name == "" && bb.Name != "" {
				bucket.Name = bb.Name
			}

			// If the description is empty, use the one from the build block
			if bucket.Description == "" && bb.Description != "" {
				bucket.Description = bb.Description
			}
			return nil
		}
	}

	// Capture Datasource configuration data
	vals, dsDiags := config.Datasources.Values()
	if dsDiags != nil {
		diags = append(diags, dsDiags...)
	}

	build := config.Builds[0]
	bucket, bucketDiags := createConfiguredBucket(
		config.Basedir,
		withPackerEnvConfiguration,
		withHCLBucketConfiguration(build),
		withDeprecatedDatasourceConfiguration(vals, ui),
		withDatasourceConfiguration(vals),
	)
	if bucketDiags != nil {
		diags = append(diags, bucketDiags...)
	}

	if diags.HasErrors() {
		return nil, diags
	}

	for _, source := range build.Sources {
		bucket.RegisterBuildForComponent(source.String())
	}

	ui.Say(fmt.Sprintf("Tracking build on HCP Packer with fingerprint %q", bucket.Version.Fingerprint))

	return &HCLRegistry{
		configuration: config,
		bucket:        bucket,
		ui:            ui,
	}, nil
}
