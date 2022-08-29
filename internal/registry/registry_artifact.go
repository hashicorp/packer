package registry

import (
	"fmt"
)

const BuilderId = "packer.post-processor.packer-registry"

type RegistryArtifact struct {
	BucketSlug  string
	IterationID string
	BuildName   string
}

func (a *RegistryArtifact) BuilderId() string {
	return BuilderId
}

func (*RegistryArtifact) Id() string {
	return ""
}

func (a *RegistryArtifact) Files() []string {
	return []string{}
}

func (a *RegistryArtifact) String() string {
	return fmt.Sprintf("Published metadata to HCP Packer registry packer/%s/iterations/%s", a.BucketSlug, a.IterationID)
}

func (*RegistryArtifact) State(name string) interface{} {
	return nil
}

func (a *RegistryArtifact) Destroy() error {
	return nil
}
