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
		// We prepend type with builder block name to prevent conflict
		name = prependIfNotEmpty(cb.BuildName, cb.Type)
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
		// We prepend type with builder block name to prevent conflict
		buildName = prependIfNotEmpty(cb.BuildName, cb.Type)
	}

	buildMetadata, envMetadata := cb.GetMetadata(), h.metadata
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

	if len(config.Builds) > 1 && config.HCPPackerRegistry == nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Multiple " + buildLabel + " blocks",
			Detail: fmt.Sprintf("For HCP Packer Registry enabled builds, only one " + buildLabel +
				" block can be defined. Please declare HCP registry configuration at root level "),
		})

		return nil, diags
	}

	withHCLBucketConfiguration := func(cfg *hcl2template.PackerConfig) bucketConfigurationOpts {
		return func(bucket *Bucket) hcl.Diagnostics {
			bucket.ReadFromHCLRoot(cfg)
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
		withHCLBucketConfiguration(config),
		withDeprecatedDatasourceConfiguration(vals, ui),
		withDatasourceConfiguration(vals),
	)
	if bucketDiags != nil {
		diags = append(diags, bucketDiags...)
	}

	if diags.HasErrors() {
		return nil, diags
	}
	buildNames := make(map[string]struct{})

	for _, build := range config.Builds {
		if config.HCPPackerRegistry != nil && build.HCPPackerRegistry != nil {
			var diags hcl.Diagnostics
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Ambiguous HCP Packer registry configuration",
				Detail: "Cannot use root declared HCP Packer configuration at the same time " +
					"as build block nested HCP Packer configuration",
			})

			return nil, diags

		}
		if build.HCPPackerRegistry != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagWarning,
				Summary:  buildLabel + " HCP registry configuration is deprecated",
				Detail:   "Please use root level HCP registry configuration",
			})

		}
		for _, source := range build.Sources {

			// We prepend each source with builder block name to prevent conflict
			buildName := prependIfNotEmpty(build.Name, source.String())
			if _, ok := buildNames[buildName]; ok {

				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Ambiguous build name",
					Detail: "build name " +
						"Two or more build blocks resolve to the same build name (" + buildName + ") " +
						"Please use the Name attribute in the build block to fix this",
				})

				return nil, diags
			}
			buildNames[buildName] = struct{}{}
			bucket.RegisterBuildForComponent(buildName)
		}
	}

	ui.Say(fmt.Sprintf("Tracking build on HCP Packer with fingerprint %q", bucket.Version.Fingerprint))

	return &HCLRegistry{
		configuration: config,
		bucket:        bucket,
		ui:            ui,
		metadata:      &MetadataStore{},
	}, nil
}

func (h *HCLRegistry) Metadata() Metadata {
	return h.metadata
}

func prependIfNotEmpty(prefix string, suffix string) string {
	if prefix != "" {
		return fmt.Sprintf("%s.%s", prefix, suffix)
	}
	return suffix
}
