package packer

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
	packerregistry "github.com/hashicorp/packer/internal/registry"
	"github.com/mitchellh/mapstructure"
)

type RegistryPostProcessor struct {
	BuilderType               string
	ArtifactMetadataPublisher *packerregistry.Bucket
	packersdk.PostProcessor
}

func (p *RegistryPostProcessor) ConfigSpec() hcldec.ObjectSpec {
	if p.PostProcessor == nil {
		return nil
	}

	return p.PostProcessor.ConfigSpec()
}

func (p *RegistryPostProcessor) Configure(raws ...interface{}) error {
	if p.PostProcessor == nil {
		return nil
	}

	return p.PostProcessor.Configure(raws...)
}

func (p *RegistryPostProcessor) PostProcess(ctx context.Context, ui packersdk.Ui, source packersdk.Artifact) (packersdk.Artifact, bool, bool, error) {
	// This is a bit of a hack for now to denote that this pp should just update the state of a build in the Packer registry.
	// TODO create an actual post-processor that we can embed here that will do the updating and printing.
	if p.PostProcessor == nil {
		if parErr := p.ArtifactMetadataPublisher.CompleteBuild(ctx, p.BuilderType); parErr != nil {
			err := fmt.Errorf("[TRACE] failed to update Packer registry with image artifacts for %q: %s", p.BuilderType, parErr)
			return nil, false, true, err
		}

		r := &RegistryArtifact{
			BuildName:   p.BuilderType,
			BucketSlug:  p.ArtifactMetadataPublisher.Slug,
			IterationID: p.ArtifactMetadataPublisher.Iteration.ID,
		}

		return r, true, false, nil
	}

	// Bump build status first so we don't end-up chaining post-processors
	// that don't heartbeat, hence letting too long happen between two
	// refreshes, and letting the build go to the FAILED status.
	err := p.ArtifactMetadataPublisher.UpdateBuildStatus(
		ctx,
		p.BuilderType,
		models.HashicorpCloudPackerBuildStatusRUNNING,
	)
	if err != nil {
		log.Printf("[TRACE] failed to heartbeat running build %s: %s", p.BuilderType, err)
	}

	cleanupHeartbeat, err := p.ArtifactMetadataPublisher.HeartbeatBuild(ctx, p.BuilderType)
	if err != nil {
		log.Printf("[ERROR] failed to start heartbeat function")
	}
	if cleanupHeartbeat != nil {
		defer cleanupHeartbeat()
	}

	source, keep, override, err := p.PostProcessor.PostProcess(ctx, ui, source)
	if err != nil {
		if parErr := p.ArtifactMetadataPublisher.UpdateBuildStatus(ctx, p.BuilderType, models.HashicorpCloudPackerBuildStatusFAILED); parErr != nil {
			log.Printf("[TRACE] failed to update Packer registry with image artifacts for %q: %s", p.BuilderType, parErr)
		}
		return source, false, false, err
	}

	var images []registryimage.Image
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &images,
		WeaklyTypedInput: true,
		ErrorUnused:      false,
	})
	if err != nil {
		return source, false, false, fmt.Errorf("failed to create decoder for HCP Packer registry image: %w", err)
	}

	state := source.State(registryimage.ArtifactStateURI)
	err = decoder.Decode(state)
	if err != nil {
		return source, false, false, fmt.Errorf("failed to obtain HCP Packer registry image from post-processor artifact: %w", err)
	}
	err = p.ArtifactMetadataPublisher.UpdateImageForBuild(p.BuilderType, images...)

	if err != nil {
		return source, keep, override, fmt.Errorf("[TRACE] failed to add image artifact for %q: %s", p.BuilderType, err)
	}

	return source, keep, override, nil
}
