package registry

import (
	"context"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	sdkpacker "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer"
)

// JSONMetadataRegistry is a HCP handler made to process legacy JSON templates
type JSONMetadataRegistry struct {
	configuration *packer.Core
	bucket        *Bucket
}

func NewJSONMetadataRegistry(config *packer.Core) (*JSONMetadataRegistry, hcl.Diagnostics) {
	bucket, diags := createConfiguredBucket(
		filepath.Dir(config.Template.Path),
		withPackerEnvConfiguration,
	)

	if diags.HasErrors() {
		return nil, diags
	}

	for _, b := range config.Template.Builders {
		// Get all builds slated within config ignoring any only or exclude flags.
		bucket.RegisterBuildForComponent(packer.HCPName(b))
	}

	return &JSONMetadataRegistry{
		configuration: config,
		bucket:        bucket,
	}, nil
}

// PopulateIteration creates the metadata on HCP for a build
func (h *JSONMetadataRegistry) PopulateIteration(ctx context.Context) error {
	for _, b := range h.configuration.Template.Builders {
		// Get all builds slated within config ignoring any only or exclude flags.
		h.bucket.RegisterBuildForComponent(b.Name)
	}

	err := h.bucket.Validate()
	if err != nil {
		return err
	}
	err = h.bucket.Initialize(ctx)
	if err != nil {
		return err
	}

	err = h.bucket.populateIteration(ctx)
	if err != nil {
		return err
	}

	return nil
}

// StartBuild is invoked when one build for the configuration is starting to be processed
func (h *JSONMetadataRegistry) StartBuild(ctx context.Context, buildName string) error {
	return h.bucket.startBuild(ctx, buildName)
}

// CompleteBuild is invoked when one build for the configuration has finished
func (h *JSONMetadataRegistry) CompleteBuild(
	ctx context.Context,
	buildName string,
	artifacts []sdkpacker.Artifact,
	buildErr error,
) ([]sdkpacker.Artifact, error) {
	return h.bucket.completeBuild(ctx, buildName, artifacts, buildErr)
}
