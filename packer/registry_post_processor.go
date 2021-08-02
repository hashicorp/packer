package packer

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	packerregistry "github.com/hashicorp/packer/internal/packer_registry"
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
		if parErr := p.ArtifactMetadataPublisher.UpdateBuildStatus(ctx, p.BuilderType, models.HashicorpCloudPackerBuildStatusDONE); parErr != nil {
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

	source, keep, override, err := p.PostProcessor.PostProcess(ctx, ui, source)
	if err != nil {
		if parErr := p.ArtifactMetadataPublisher.UpdateBuildStatus(ctx, p.BuilderType, models.HashicorpCloudPackerBuildStatusFAILED); parErr != nil {
			log.Printf("[TRACE] failed to update Packer registry with image artifacts for %q: %s", p.BuilderType, parErr)
		}
		return source, false, false, err
	}

	// Lets post state
	if source != nil {
		switch state := source.State("par.artifact.metadata").(type) {
		case map[interface{}]interface{}:
			m := make(map[string]string)
			for k, v := range state {
				m[k.(string)] = v.(string)
			}

			// TODO handle these error better
			err := p.ArtifactMetadataPublisher.UpdateImageForBuild(p.BuilderType, packerregistry.Image{
				ProviderName:   m["ProviderName"],
				ProviderRegion: m["ProviderRegion"],
				ID:             m["ImageID"],
			})
			if err != nil {
				log.Printf("[TRACE] failed to add image artifact for %q: %s", p.BuilderType, err)
			}
		case []interface{}:
			for _, d := range state {
				d := d.(map[interface{}]interface{})
				err := p.ArtifactMetadataPublisher.UpdateImageForBuild(p.BuilderType, packerregistry.Image{
					ProviderName:   d["ProviderName"].(string),
					ProviderRegion: d["ProviderRegion"].(string),
					ID:             d["ImageID"].(string),
				})
				if err != nil {
					log.Printf("[TRACE] failed to add image artifact for %q: %s", p.BuilderType, err)
				}
			}
		}
	}

	return source, keep, override, nil
}
