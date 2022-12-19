package registry

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
	sdkpacker "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer"
)

// JSONMetadataRegistry is a HCP handler made to process legacy JSON templates
type JSONMetadataRegistry struct {
	configuration *packer.Core
	bucket        *Bucket
	ui            sdkpacker.Ui
}

func NewJSONMetadataRegistry(config *packer.Core, ui sdkpacker.Ui) (*JSONMetadataRegistry, hcl.Diagnostics) {
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

	ui.Say(fmt.Sprintf("Tracking build on HCP Packer with fingerprint %q", bucket.Iteration.Fingerprint))

	return &JSONMetadataRegistry{
		configuration: config,
		bucket:        bucket,
		ui:            ui,
	}, nil
}

// PopulateIteration creates the metadata on HCP for a build
func (h *JSONMetadataRegistry) PopulateIteration(ctx context.Context) error {
	err := h.bucket.Validate()
	if err != nil {
		return err
	}
	err = h.bucket.Initialize(ctx, models.HashicorpCloudPackerIterationTemplateTypeJSON)
	if err != nil {
		return err
	}

	err = h.bucket.populateIteration(ctx)
	if err != nil {
		return err
	}

	sha, err := getGitSHA(h.configuration.Template.Path)
	if err != nil {
		log.Printf("failed to get GIT SHA from environment, won't set as build labels")
	} else {
		h.bucket.Iteration.AddSHAToBuildLabels(sha)
	}

	return nil
}

// StartBuild is invoked when one build for the configuration is starting to be processed
func (h *JSONMetadataRegistry) StartBuild(ctx context.Context, build sdkpacker.Build) error {
	return h.bucket.startBuild(ctx, build.Name())
}

// CompleteBuild is invoked when one build for the configuration has finished
func (h *JSONMetadataRegistry) CompleteBuild(
	ctx context.Context,
	build sdkpacker.Build,
	artifacts []sdkpacker.Artifact,
	buildErr error,
) ([]sdkpacker.Artifact, error) {
	return h.bucket.completeBuild(ctx, build.Name(), artifacts, buildErr)
}

// IterationStatusSummary prints a status report in the UI if the iteration is not yet done
func (h *JSONMetadataRegistry) IterationStatusSummary() {
	h.bucket.Iteration.iterationStatusSummary(h.ui)
}
