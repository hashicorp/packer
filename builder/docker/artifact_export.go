package docker

import (
	"fmt"
	"os"
)

// ExportArtifact is an Artifact implementation for when a container is
// exported from docker into a single flat file.
type ExportArtifact struct {
	path string
}

func (*ExportArtifact) BuilderId() string {
	return BuilderId
}

func (a *ExportArtifact) Files() []string {
	return []string{a.path}
}

func (*ExportArtifact) Id() string {
	return "Container"
}

func (a *ExportArtifact) String() string {
	return fmt.Sprintf("Exported Docker file: %s", a.path)
}

func (a *ExportArtifact) State(name string) interface{} {
	return nil
}

func (a *ExportArtifact) Destroy() error {
	return os.Remove(a.path)
}
