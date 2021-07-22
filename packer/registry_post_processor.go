package packer

import (
	"context"
	"log"
	"strings"

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
	if p.PostProcessor == nil {
		if parErr := p.ArtifactMetadataPublisher.PublishBuildStatus(ctx, p.BuilderType, models.HashicorpCloudPackerBuildStatusDONE); parErr != nil {
			log.Printf("[TRACE] failed to update Packer registry with image artifacts for %q: %s", p.BuilderType, parErr)
		}
		return nil, true, false, nil
	}

	source, keep, override, err := p.PostProcessor.PostProcess(ctx, ui, source)
	if err != nil {
		if parErr := p.ArtifactMetadataPublisher.PublishBuildStatus(ctx, p.BuilderType, models.HashicorpCloudPackerBuildStatusFAILED); parErr != nil {
			log.Printf("[TRACE] failed to update Packer registry with image artifacts for %q: %s", p.BuilderType, parErr)
		}
		return source, false, false, err
	}

	// Lets post state
	if source != nil {
		metadata := make(map[string]string)
		metadata[source.BuilderId()] = source.String()
		if len(source.Files()) > 0 {
			metadata[source.BuilderId()+".files"] = strings.Join(source.Files(), ", ")
		}
		err := p.ArtifactMetadataPublisher.AddBuildMetadata(p.BuilderType, metadata)
		if err != nil {
			log.Printf("[TRACE] failed to add build labels for %q: %s", p.BuilderType, err)
		}

		switch state := source.State("par.artifact.metadata").(type) {
		case map[interface{}]interface{}:
			m := make(map[string]string)
			for k, v := range state {
				m[k.(string)] = v.(string)
			}

			// TODO handle these error better
			err := p.ArtifactMetadataPublisher.AddBuildArtifact(p.BuilderType, packerregistry.PARtifact{
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
				err := p.ArtifactMetadataPublisher.AddBuildArtifact(p.BuilderType, packerregistry.PARtifact{
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
