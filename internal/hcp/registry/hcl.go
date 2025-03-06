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
	metadata      *MetadataStore
	buildNames    map[string]struct{}
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
func (h *HCLRegistry) StartBuild(ctx context.Context, build *packer.CoreBuild) error {
	return h.bucket.startBuild(ctx, h.HCPBuildName(build))
}

// CompleteBuild is invoked when one build for the configuration has finished
func (h *HCLRegistry) CompleteBuild(
	ctx context.Context,
	build *packer.CoreBuild,
	artifacts []sdkpacker.Artifact,
	buildErr error,
) ([]sdkpacker.Artifact, error) {
	buildName := h.HCPBuildName(build)
	buildMetadata, envMetadata := build.GetMetadata(), h.metadata
	err := h.bucket.Version.AddMetadataToBuild(ctx, buildName, buildMetadata, envMetadata)
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

	registryConfig, rcDiags := config.GetHCPPackerRegistryBlock()
	diags = diags.Extend(rcDiags)
	if diags.HasErrors() {
		return nil, diags
	}

	withHCLBucketConfiguration := func(bucket *Bucket) hcl.Diagnostics {
		bucket.ReadFromHCPPackerRegistryBlock(registryConfig)
		return nil
	}

	// we must use the old strategy when there is only a single build block because
	// we used to rely on the parent build block for setting some default data
	if len(config.Builds) == 1 && config.HCPPackerRegistry == nil {
		withHCLBucketConfiguration = func(bucket *Bucket) hcl.Diagnostics {
			bb := config.Builds[0]
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

	bucket, bucketDiags := createConfiguredBucket(
		config.Basedir,
		withPackerEnvConfiguration,
		withHCLBucketConfiguration,
		withDeprecatedDatasourceConfiguration(vals, ui),
		withDatasourceConfiguration(vals),
	)
	if bucketDiags != nil {
		diags = append(diags, bucketDiags...)
	}

	if diags.HasErrors() {
		return nil, diags
	}

	registry := &HCLRegistry{
		configuration: config,
		bucket:        bucket,
		ui:            ui,
		metadata:      &MetadataStore{},
		buildNames:    map[string]struct{}{},
	}

	ui.Say(fmt.Sprintf("Tracking build on HCP Packer with fingerprint %q", bucket.Version.Fingerprint))

	return registry, diags.Extend(registry.registerAllComponents())
}

func (h *HCLRegistry) registerAllComponents() hcl.Diagnostics {
	var diags hcl.Diagnostics

	conflictSources := map[string]struct{}{}

	// we currently support only one build block but it will change in the near future
	for _, build := range h.configuration.Builds {
		for _, source := range build.Sources {
			// If we encounter the same source twice, we'll defer
			// its addition to later, using both the build name
			// and the source type as the name used for HCP Packer.
			_, ok := h.buildNames[source.String()]
			if !ok {
				h.buildNames[source.String()] = struct{}{}
				continue
			}

			conflictSources[source.String()] = struct{}{}
			// We need to delete it to avoid having a false-positive
			// when returning the name, since we'll be using
			// the combination of build name + source.String()
			delete(h.buildNames, source.String())
		}
	}

	// Second pass is to take care of conflicting sources
	//
	// If the same source is used twice in the configuration, we need to
	// have a way to differentiate the two on HCP, as each build should have
	// a locally unique name.
	//
	// If that happens, we then use a combination of both the build name, and
	// the source type.
	for _, build := range h.configuration.Builds {
		for _, source := range build.Sources {
			if _, ok := conflictSources[source.String()]; !ok {
				continue
			}

			buildName := source.String()
			if build.Name != "" {
				buildName = fmt.Sprintf("%s.%s", build.Name, buildName)
			}

			if _, ok := h.buildNames[buildName]; ok {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Build name conflicts",
					Subject:  &build.HCL2Ref.DefRange,
					Detail: fmt.Sprintf("Two sources are used in the same build block, causing "+
						"a conflict, there must only be one instance of %s", source.String()),
				})
			}
			h.buildNames[buildName] = struct{}{}
		}
	}

	if diags.HasErrors() {
		return diags
	}

	for buildName := range h.buildNames {
		h.bucket.RegisterBuildForComponent(buildName)
	}
	return diags
}

func (h *HCLRegistry) Metadata() Metadata {
	return h.metadata
}

// HCPBuildName will return the properly formatted string taking name conflict into account
func (h *HCLRegistry) HCPBuildName(build *packer.CoreBuild) string {
	_, ok := h.buildNames[build.Type]
	if ok {
		return build.Type
	}

	return fmt.Sprintf("%s.%s", build.BuildName, build.Type)
}
