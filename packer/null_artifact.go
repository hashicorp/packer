package packer

import (
	"fmt"
)

// dummy Artifact implementation - does nothing
type NullArtifact struct {
	BuilderIdValue string
}

func (a *NullArtifact) BuilderId() string {
	return a.BuilderIdValue
}

func (a *NullArtifact) Files() []string {
	return []string{}
}

func (*NullArtifact) Id() string {
	return "Null"
}

func (a *NullArtifact) String() string {
	return fmt.Sprintf("Did not export anything.")
}

func (a *NullArtifact) State(name string) interface{} {
	return nil
}

func (a *NullArtifact) Destroy() error {
	return nil
}
