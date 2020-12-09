package docker

import (
	"fmt"
	"os"
)

// ExportArtifact is an Artifact implementation for when a container is
// exported from docker into a single flat file.
type ExportArtifact struct {
	path string
	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
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
	return a.StateData[name]
}

func (a *ExportArtifact) Destroy() error {
	if a.path != "" {
		return os.Remove(a.path)
	}
	return nil
}
