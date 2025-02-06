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
	ui            sdkpacker.Ui
	metadata      *MetadataStore

	bucketsByName      map[string]*Bucket
	bucketsByBuildName map[string]*Bucket
}

const (
	// Known HCP Packer Datasource, whose id is the SourceImageId for some build.
	hcpImageDatasourceType    string = "hcp-packer-image"
	hcpArtifactDatasourceType string = "hcp-packer-artifact"

	hcpIterationDatasourceType string = "hcp-packer-iteration"
	hcpVersionDatasourceType   string = "hcp-packer-version"

	buildLabel string = "build"
)

// PopulateVersion creates the metadata in HCP Packer Registry for a build on all buckets
func (h *HCLRegistry) PopulateVersion(ctx context.Context) error {
	for _, bucket := range h.bucketsByName {
		err := bucket.Initialize(ctx, hcpPackerModels.HashicorpCloudPacker20230101TemplateTypeHCL2)
		if err != nil {
			return err
		}

		err = bucket.populateVersion(ctx)
		if err != nil {
			return err
		}

		versionID := bucket.Version.ID
		versionFingerprint := bucket.Version.Fingerprint

		// FIXME: Remove
		h.configuration.HCPVars["iterationID"] = cty.StringVal(versionID)
		h.configuration.HCPVars["versionFingerprint"] = cty.StringVal(versionFingerprint)

		sha, err := getGitSHA(h.configuration.Basedir)
		if err != nil {
			log.Printf("failed to get GIT SHA from environment, won't set as build labels")
		} else {
			bucket.Version.AddSHAToBuildLabels(sha)
		}
	}

	return nil
}

// StartBuild is invoked when one build for the configuration is starting to be processed
func (h *HCLRegistry) StartBuild(ctx context.Context, build sdkpacker.Build) error {
	name := build.Name()
	cb, ok := build.(*packer.CoreBuild)
	if ok {
		// We prepend type with builder block name to prevent conflict
		name = fmt.Sprintf("%s.%s", cb.BuildName, cb.Type)
	}

	return h.bucketsByBuildName[name].startBuild(ctx, name)
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
		// We prepend type with builder block name to prevent conflict
		buildName = fmt.Sprintf("%s.%s", cb.BuildName, cb.Type)
	}

	buildMetadata, envMetadata := cb.GetMetadata(), h.metadata
	err := h.bucketsByBuildName[buildName].Version.AddMetadataToBuild(ctx, buildName, buildMetadata, envMetadata)
	if err != nil {
		return nil, err
	}
	return h.bucketsByBuildName[buildName].completeBuild(ctx, buildName, artifacts, buildErr)
}

// VersionStatusSummary prints a status report for each bucket in the UI if the version is not yet done
func (h *HCLRegistry) VersionStatusSummary() {
	for _, bucket := range h.bucketsByName {
		bucket.Version.statusSummary(h.ui)
	}
}

func NewHCLRegistry(config *hcl2template.PackerConfig, ui sdkpacker.Ui) (*HCLRegistry, hcl.Diagnostics) {
	var diags hcl.Diagnostics

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

	bucketsByName := map[string]*Bucket{}
	for _, build := range config.Builds {
		bucketName := build.HCPPackerRegistry.Slug
		if _, ok := bucketsByName[bucketName]; ok {
			continue
		}

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
		bucketsByName[bucketName] = bucket
		ui.Say(fmt.Sprintf(
			"Tracking build on HCP Packer bucket '%s' with fingerprint %q",
			bucketName,
			bucket.Version.Fingerprint,
		))
	}

	bucketsByBuildName := map[string]*Bucket{}
	for _, build := range config.Builds {
		for _, source := range build.Sources {

			// We prepend each source with builder block name to prevent conflict
			buildName := fmt.Sprintf("%s.%s", build.Name, source.String())
			bucketsByBuildName[buildName] = bucketsByName[build.HCPPackerRegistry.Slug]
			bucketsByBuildName[buildName].RegisterBuildForComponent(buildName)
		}
	}

	return &HCLRegistry{
		configuration: config,
		ui:            ui,
		metadata:      &MetadataStore{},

		bucketsByName:      bucketsByName,
		bucketsByBuildName: bucketsByBuildName,
	}, nil
}

func (h *HCLRegistry) Metadata() Metadata {
	return h.metadata
}
