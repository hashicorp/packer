package registry

import (
	"runtime"
)

// Metadata is the global metadata store, it is attached to a registry implementation
// and keeps track of the environmental information.
// This then can be sent to HCP Packer, so we can present it to users.
type Metadata interface {
	// Gather is the point where we vacuum all the information
	// relevant from the environment in order to expose it to HCP Packer.
	Gather(args map[string]interface{}) error
	// Render is called when the metadata is sent to HCP Packer,
	// i.e. when a build finishes, so it gets merged with the rest of the
	// information that is build-specific
	Render() map[string]interface{}
}

// MetadataStore is the effective implementation of a global store for metadata
// destined to be uploaded to HCP Packer.
//
// If HCP is enabled during a build, this is populated with a curated list of
// arguments to the build command, and environment-related information.
type MetadataStore struct {
	PackerBuildCommandOptions map[string]interface{}
	OperatingSystem           map[string]string
}

func (ms *MetadataStore) Gather(args map[string]interface{}) error {
	// Environment data
	ms.gatherOperatingSystemDetails()

	// Build arguments
	ms.PackerBuildCommandOptions = args

	return nil
}

func (ms *MetadataStore) gatherOperatingSystemDetails() {
	ms.OperatingSystem = map[string]string{
		"os":   runtime.GOOS,
		"arch": runtime.GOARCH,
	}
}

func (ms *MetadataStore) Render() map[string]interface{} {
	return map[string]interface{}{
		"operating_system":             ms.OperatingSystem,
		"packer_build_command_options": ms.PackerBuildCommandOptions,
	}
}

// NilMetadata is a dummy implementation of a Metadata that does nothing.
//
// It is the implementation used typically when HCP is disabled, so nothing is
// collected or kept in memory in this case.
type NilMetadata struct{}

func (ns NilMetadata) Gather(args map[string]interface{}) error {
	return nil
}

func (ns NilMetadata) Render() map[string]interface{} {
	return map[string]interface{}{}
}
