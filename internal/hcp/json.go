package hcp

import (
	"context"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	sdkpacker "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/internal/registry"
	"github.com/hashicorp/packer/internal/registry/env"
	"github.com/hashicorp/packer/packer"
)

// jsonOrchestrator is a HCP handler made to process legacy JSON templates
type jsonOrchestrator struct {
	configuration *packer.Core
	bucket        *registry.Bucket
}

func newJSONOrchestrator(config *packer.Core) (Orchestrator, hcl.Diagnostics) {
	if env.IsHCPDisabled() ||
		(!env.HasPackerRegistryBucket() && !env.IsHCPExplicitelyEnabled()) {
		return newNoopHandler(), nil
	}

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

	return &jsonOrchestrator{
		configuration: config,
		bucket:        bucket,
	}, nil
}

// PopulateIteration creates the metadata on HCP for a build
func (h *jsonOrchestrator) PopulateIteration(ctx context.Context) error {
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

	err = h.bucket.PopulateIteration(ctx)
	if err != nil {
		return err
	}

	return nil
}

// BuildStart is invoked when one build for the configuration is starting to be processed
func (h *jsonOrchestrator) BuildStart(ctx context.Context, buildName string) error {
	return h.bucket.BuildStart(ctx, buildName)
}

// BuildDone is invoked when one build for the configuration has finished
func (h *jsonOrchestrator) BuildDone(
	ctx context.Context,
	buildName string,
	artifacts []sdkpacker.Artifact,
	buildErr error,
) ([]sdkpacker.Artifact, error) {
	return h.bucket.BuildDone(ctx, buildName, artifacts, buildErr)
}
