package vagrant

import (
	"github.com/hashicorp/packer/packer"
)

// Provider is the interface that each provider must implement in order
// to package the artifacts into a Vagrant-compatible box.
type Provider interface {
	// KeepInputArtifact should return true/false whether this provider
	// requires the input artifact to be kept by default.
	KeepInputArtifact() bool

	// Process is called to process an artifact into a Vagrant box. The
	// artifact is given as well as the temporary directory path to
	// put things.
	//
	// The Provider should return the contents for the Vagrantfile,
	// any metadata (including the provider type in that), and an error
	// if any.
	Process(packer.Ui, packer.Artifact, string) (vagrantfile string, metadata map[string]interface{}, err error)
}
