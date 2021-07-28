package packer_registry

import (
	"sync"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
)

// Build represents a build of a given component type for some bucket on the HCP Packer Registry.
type Build struct {
	ID            string
	CloudProvider string
	ComponentType string
	RunUUID       string
	Metadata      map[string]string
	Images        []Image
	Status        models.HashicorpCloudPackerBuildStatus
}

// Builds a concurrent safe collection of Build entries that can be associated to some Iteration.
type Builds struct {
	sync.RWMutex
	m map[string]*Build
}

// Image represents an artifact on some external provider (e.g AWS, GCP, Azure) that should be tracked
// as the main image artifact for some iteration of a Bucket on the HCP Packer Registry.
type Image struct {
	ID                           string
	ProviderName, ProviderRegion string
}
