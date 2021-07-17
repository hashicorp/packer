package packer

import (
	"context"
	"log"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	packerregistry "github.com/hashicorp/packer/internal/packer_registry"
)

type RegistryBuilder struct {
	ArtifactMetadataPublisher *packerregistry.Bucket
	Name                      string
	packersdk.Builder
}

func (b *RegistryBuilder) Prepare(raws ...interface{}) ([]string, []string, error) {
	return b.Builder.Prepare(raws)
}

// Run is where the actual build should take place. It takes a Build and a Ui.
func (b *RegistryBuilder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) {
	if ctx.Err() != nil { // context was cancelled
		if err := b.ArtifactMetadataPublisher.PublishBuildStatus(ctx, b.Name, models.HashicorpCloudPackerBuildStatusCANCELLED); err != nil {
			log.Printf("[TRACE] failed to update Packer registry status for %q: %s", b.Name, err)
		}
		return nil, ctx.Err()
	}
	if err := b.ArtifactMetadataPublisher.PublishBuildStatus(ctx, b.Name, models.HashicorpCloudPackerBuildStatusRUNNING); err != nil {
		log.Printf("[TRACE] failed to update Packer registry status for %q: %s", b.Name, err)
	}

	artifact, err := b.Builder.Run(ctx, ui, hook)
	if err != nil {
		if parErr := b.ArtifactMetadataPublisher.PublishBuildStatus(ctx, b.Name, models.HashicorpCloudPackerBuildStatusFAILED); parErr != nil {
			log.Printf("[TRACE] failed to update Packer registry status for %q: %s", b.Name, parErr)
		}
	}

	return artifact, err
}
