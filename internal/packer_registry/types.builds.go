package packer_registry

import (
	"sync"

	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
)

type Build struct {
	ID            string
	CloudProvider string
	ComponentType string
	RunUUID       string
	Metadata      map[string]string
	PARtifacts    []PARtifact
	Status        models.HashicorpCloudPackerBuildStatus
}

type Builds struct {
	sync.RWMutex
	m map[string]*Build
}

func NewBuilds() Builds {
	return Builds{
		m: make(map[string]*Build),
	}
}

type PARtifact struct {
	ID                           string
	ProviderName, ProviderRegion string
}
