package registry

import (
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
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
