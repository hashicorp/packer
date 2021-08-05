package packer

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	packerregistry "github.com/hashicorp/packer/internal/packer_registry"
)

type RegistryBuilder struct {
	Name                      string
	ArtifactMetadataPublisher *packerregistry.Bucket
	packersdk.Builder
}

func (b *RegistryBuilder) Prepare(raws ...interface{}) ([]string, []string, error) {
	return b.Builder.Prepare(raws...)
}

// Run is where the actual build should take place. It takes a Build and a Ui.
func (b *RegistryBuilder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) {

	if !b.ArtifactMetadataPublisher.IsExpectingBuildForComponent(b.Name) {
		ui.Error(fmt.Sprintf("The build for %q in iteration %q has already been marked as DONE; Skipping build to prevent drift.", b.Name, b.ArtifactMetadataPublisher.Iteration.ID))
		return nil, nil
	}

	runCompleted := make(chan struct{})
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Printf("[TRACE] marking build %q as cancelled in HCP Packer registry", b.Name)
				if err := b.ArtifactMetadataPublisher.UpdateBuildStatus(context.TODO(), b.Name, models.HashicorpCloudPackerBuildStatusCANCELLED); err != nil {
					log.Printf("[TRACE] failed to update HCP Packer registry status for %q: %s", b.Name, err)
				}
				return
			case <-runCompleted:
				return
			}
		}
	}()

	if err := b.ArtifactMetadataPublisher.UpdateBuildStatus(ctx, b.Name, models.HashicorpCloudPackerBuildStatusRUNNING); err != nil {
		log.Printf("[TRACE] failed to update HCP Packer registry status for %q: %s", b.Name, err)
	}

	ui.Say(fmt.Sprintf("Publishing build details for %s to the HCP Packer registry", b.Name))
	artifact, err := b.Builder.Run(ctx, ui, hook)
	if err != nil {
		if parErr := b.ArtifactMetadataPublisher.UpdateBuildStatus(ctx, b.Name, models.HashicorpCloudPackerBuildStatusFAILED); parErr != nil {
			log.Printf("[TRACE] failed to update HCP Packer registry status for %q: %s", b.Name, parErr)
		}
	}

	// close chan to mark completion
	close(runCompleted)

	// Lets post state
	if artifact != nil {
		switch state := artifact.State("par.artifact.metadata").(type) {
		case map[interface{}]interface{}:
			m := make(map[string]string)
			for k, v := range state {
				m[k.(string)] = v.(string)
			}
			// TODO handle these error better
			err := b.ArtifactMetadataPublisher.UpdateImageForBuild(b.Name, packerregistry.Image{
				ProviderName:   m["ProviderName"],
				ProviderRegion: m["ProviderRegion"],
				ID:             m["ImageID"],
			})
			if err != nil {
				log.Printf("[TRACE] failed to add image artifact for %q: %s", b.Name, err)
			}
		case []interface{}:
			for _, d := range state {
				d := d.(map[interface{}]interface{})
				err := b.ArtifactMetadataPublisher.UpdateImageForBuild(b.Name, packerregistry.Image{
					ProviderName:   d["ProviderName"].(string),
					ProviderRegion: d["ProviderRegion"].(string),
					ID:             d["ImageID"].(string),
				})
				if err != nil {
					log.Printf("[TRACE] failed to add image artifact for %q: %s", b.Name, err)
				}
			}
		}
	}

	return artifact, err
}
