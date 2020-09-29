package docker

import (
	"fmt"
	"strings"
)

// ImportArtifact is an Artifact implementation for when a container is
// exported from docker into a single flat file.
type ImportArtifact struct {
	BuilderIdValue string
	Driver         Driver
	IdValue        string

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
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
	var tags []string
	switch t := a.StateData["docker_tags"].(type) {
	case []string:
		tags = t
	case []interface{}:
		for _, name := range t {
			if n, ok := name.(string); ok {
				tags = append(tags, n)
			}
		}
	}
	if len(tags) > 0 {
		return fmt.Sprintf("Imported Docker image: %s with tags %s",
			a.Id(), strings.Join(tags, " "))
	}
	return fmt.Sprintf("Imported Docker image: %s", a.Id())
}

func (a *ImportArtifact) State(name string) interface{} {
	return a.StateData[name]
}

func (a *ImportArtifact) Destroy() error {
	return a.Driver.DeleteImage(a.Id())
}
