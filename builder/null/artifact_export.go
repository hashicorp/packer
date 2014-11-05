package null

import (
	"fmt"
)

// dummy Artifact implementation - does nothing
type NullArtifact struct {
}

func (*NullArtifact) BuilderId() string {
	return BuilderId
}

func (a *NullArtifact) Files() []string {
	return []string{}
}

func (*NullArtifact) Id() string {
	return "Null"
}

func (a *NullArtifact) String() string {
	return fmt.Sprintf("Did not export anything. This is the null builder")
}

func (a *NullArtifact) State(name string) interface{} {
	return nil
}

func (a *NullArtifact) Destroy() error {
	return nil
}
