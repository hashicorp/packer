package registry

import (
	"fmt"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
)

// Build represents a build of a given component type for some bucket on the HCP Packer Registry.
type Build struct {
	ID            string
	CloudProvider string
	ComponentType string
	RunUUID       string
	Labels        map[string]string
	Images        map[string]registryimage.Image
	Status        models.HashicorpCloudPackerBuildStatus
}

// NewBuildFromCloudPackerBuild converts a HashicorpCloudePackerBuild to a local build that can be tracked and published to the HCP Packer Registry.
// Any existing labels or images associated to src will be copied to the returned Build.
func NewBuildFromCloudPackerBuild(src *models.HashicorpCloudPackerBuild) (*Build, error) {

	build := Build{
		ID:            src.ID,
		ComponentType: src.ComponentType,
		CloudProvider: src.CloudProvider,
		RunUUID:       src.PackerRunUUID,
		Status:        src.Status,
		Labels:        src.Labels,
	}

	var err error
	for _, image := range src.Images {
		image := image
		err = build.AddImages(registryimage.Image{
			ImageID:        image.ImageID,
			ProviderName:   build.CloudProvider,
			ProviderRegion: image.Region,
		})

		if err != nil {
			return nil, fmt.Errorf("NewBuildFromCloudPackerBuild: %w", err)
		}
	}

	return &build, nil
}

// AddLabelsToBuild merges the contents of data to the labels associated with the build.
// Duplicate keys will be updated to reflect the new value.
func (b *Build) MergeLabels(data map[string]string) {
	if data == nil {
		return
	}

	if b.Labels == nil {
		b.Labels = make(map[string]string)
	}

	for k, v := range data {
		// TODO: (nywilken) Determine why we skip labels already set
		//if _, ok := build.Labels[k]; ok {
		//continue
		//}
		b.Labels[k] = v
	}

}

// AddImages appends one or more images artifacts to the build.
func (b *Build) AddImages(images ...registryimage.Image) error {

	if b.Images == nil {
		b.Images = make(map[string]registryimage.Image)
	}

	for _, image := range images {
		image := image

		if err := image.Validate(); err != nil {
			return fmt.Errorf("AddImages: failed to add image to build %q: %w", b.ComponentType, err)
		}

		if b.CloudProvider == "" {
			b.CloudProvider = image.ProviderName
		}

		b.MergeLabels(image.Labels)
		b.Images[image.String()] = image
	}

	return nil
}

// IsNotDone returns true if build does not satisfy all requirements of a completed build.
// A completed build must have a valid ID, one or more Images, and its Status is HashicorpCloudPackerBuildStatusDONE.
func (b *Build) IsNotDone() bool {
	hasBuildID := b.ID != ""
	hasNoImages := len(b.Images) == 0
	isNotDone := b.Status != models.HashicorpCloudPackerBuildStatusDONE

	return hasBuildID && hasNoImages && isNotDone
}
