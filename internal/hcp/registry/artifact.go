package registry

import (
	"fmt"
)

const BuilderId = "packer.post-processor.hpc-packer-registry"

type registryArtifact struct {
	BucketSlug  string
	IterationID string
	BuildName   string
}

func (a *registryArtifact) BuilderId() string {
	return BuilderId
}

func (*registryArtifact) Id() string {
	return ""
}

func (a *registryArtifact) Files() []string {
	return []string{}
}

func (a *registryArtifact) String() string {
	return fmt.Sprintf("Published metadata to HCP Packer registry packer/%s/iterations/%s", a.BucketSlug, a.IterationID)
}

func (*registryArtifact) State(name string) interface{} {
	return nil
}

func (a *registryArtifact) Destroy() error {
	return nil
}
