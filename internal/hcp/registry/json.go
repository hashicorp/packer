// Copyright IBM Corp. 2013, 2025
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
	"github.com/hashicorp/packer/hcl2template"
	"github.com/hashicorp/packer/packer"
)

// JSONRegistry is a HCP handler made to process legacy JSON templates
type JSONRegistry struct {
	configuration *packer.Core
	bucket        *Bucket
	ui            sdkpacker.Ui
	metadata      *MetadataStore
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
		metadata:      &MetadataStore{},
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
func (h *JSONRegistry) StartBuild(ctx context.Context, build *packer.CoreBuild) error {
	return h.bucket.startBuild(ctx, build.Name())
}

// CompleteBuild is invoked when one build for the configuration has finished
func (h *JSONRegistry) CompleteBuild(
	ctx context.Context,
	build *packer.CoreBuild,
	artifacts []sdkpacker.Artifact,
	buildErr error,
) ([]sdkpacker.Artifact, error) {
	buildName := build.Name()
	buildMetadata, envMetadata := build.GetMetadata(), h.metadata
	err := h.bucket.Version.AddMetadataToBuild(ctx, buildName, buildMetadata, envMetadata)
	if err != nil {
		return nil, err
	}
	return h.bucket.completeBuild(ctx, buildName, artifacts, h.ui, buildErr)
}

// VersionStatusSummary prints a status report in the UI if the version is not yet done
func (h *JSONRegistry) VersionStatusSummary() {
	h.bucket.Version.statusSummary(h.ui)
}

// Metadata gets the global metadata object that registers global settings
func (h *JSONRegistry) Metadata() Metadata {
	return h.metadata
}

// FetchEnforcedBlocks fetches enforced provisioner blocks from HCP Packer
func (h *JSONRegistry) FetchEnforcedBlocks(ctx context.Context) error {
	return h.bucket.FetchEnforcedBlocks(ctx)
}

// InjectEnforcedProvisioners injects enforced provisioners into the builds
func (h *JSONRegistry) InjectEnforcedProvisioners(builds []*packer.CoreBuild) hcl.Diagnostics {
	enforcedBlocks := h.bucket.EnforcedBlocks
	if len(enforcedBlocks) == 0 {
		return nil
	}

	var allDiags hcl.Diagnostics

	for _, eb := range enforcedBlocks {
		if eb.BlockContent == "" {
			continue
		}

		provBlocks, diags := hcl2template.ParseProvisionerBlocks(eb.BlockContent)
		if diags.HasErrors() {
			allDiags = append(allDiags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Failed to parse enforced block %q", eb.Name),
				Detail:   diags.Error(),
			})
			continue
		}

		if len(provBlocks) > 0 {
			h.ui.Say(fmt.Sprintf("Loaded %d enforced provisioner(s) from HCP block %q and template type %q", len(provBlocks), eb.Name, eb.TemplateType))
		}

		for _, build := range builds {
			buildName := build.Type
			injected := make([]packer.CoreBuildProvisioner, 0, len(provBlocks))

			for _, pb := range provBlocks {
				if pb.OnlyExcept.Skip(buildName) {
					log.Printf("[DEBUG] skipping enforced provisioner %q for legacy JSON build %q due to only/except rules",
						pb.PType, build.Name())
					continue
				}

				coreProv, moreDiags := h.configuration.GenerateCoreBuildProvisionerFromHCLBody(
					pb.PType,
					pb.Rest,
					pb.Override,
					pb.PauseBefore,
					pb.MaxRetries,
					pb.Timeout,
					buildName,
				)
				if moreDiags.HasErrors() {
					allDiags = append(allDiags, moreDiags...)
					continue
				}

				build.Provisioners = append(build.Provisioners, coreProv)
				injected = append(injected, coreProv)

				log.Printf("[INFO] injected enforced provisioner %q from block %q into legacy JSON build %q",
					pb.PType, eb.Name, build.Name())
			}

			if len(injected) == 0 {
				continue
			}

			if err := build.PrepareProvisioners(injected...); err != nil {
				allDiags = append(allDiags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  fmt.Sprintf("Failed to prepare enforced provisioners for legacy JSON build %q", build.Name()),
					Detail:   err.Error(),
				})
			}
		}
	}

	return allDiags
}
