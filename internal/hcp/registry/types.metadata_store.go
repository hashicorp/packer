package registry

import "github.com/hashicorp/packer/internal/hcp/registry/metadata"

// Metadata is the global metadata store, it is attached to a registry implementation
// and keeps track of the environmental information.
// This then can be sent to HCP Packer, so we can present it to users.
type Metadata interface {
	// Gather is the point where we vacuum all the information
	// relevant from the environment in order to expose it to HCP Packer.
	Gather(args map[string]interface{})
}

// MetadataStore is the effective implementation of a global store for metadata
// destined to be uploaded to HCP Packer.
//
// If HCP is enabled during a build, this is populated with a curated list of
// arguments to the build command, and environment-related information.
type MetadataStore struct {
	PackerBuildCommandOptions map[string]interface{}
	OperatingSystem           map[string]interface{}
	Vcs                       map[string]interface{}
	Cicd                      map[string]interface{}
}

func (ms *MetadataStore) Gather(args map[string]interface{}) {
	ms.OperatingSystem = metadata.GetOSMetadata()
	ms.Cicd = metadata.GetCicdMetadata()
	ms.Vcs = metadata.GetVcsMetadata()
	ms.PackerBuildCommandOptions = args
}

// NilMetadata is a dummy implementation of a Metadata that does nothing.
//
// It is the implementation used typically when HCP is disabled, so nothing is
// collected or kept in memory in this case.
type NilMetadata struct{}

func (ns NilMetadata) Gather(args map[string]interface{}) {}
