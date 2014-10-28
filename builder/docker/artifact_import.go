package docker

import (
	"fmt"
)

// ImportArtifact is an Artifact implementation for when a container is
// exported from docker into a single flat file.
type ImportArtifact struct {
	BuilderIdValue string
	Driver         Driver
	IdValue        string
}

func (a *ImportArtifact) BuilderId() string {
	return a.BuilderIdValue
}

func (*ImportArtifact) Files() []string {
	return nil
}

func (a *ImportArtifact) Id() string {
	return a.IdValue
}

func (a *ImportArtifact) String() string {
	return fmt.Sprintf("Imported Docker image: %s", a.Id())
}

func (*ImportArtifact) State(name string) interface{} {
	return nil
}

func (a *ImportArtifact) Destroy() error {
	return a.Driver.DeleteImage(a.Id())
}
