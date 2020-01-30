package oneandone

import (
	"fmt"
)

type Artifact struct {
	snapshotId   string
	snapshotName string

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return []string{}
}

func (*Artifact) Id() string {
	return "Null"
}

func (a *Artifact) String() string {
	if a.snapshotId == "" {
		return "No image has been created."
	}
	return fmt.Sprintf("A snapshot was created: '%v', '%v'", a.snapshotId, a.snapshotName)
}

func (a *Artifact) State(name string) interface{} {
	return a.StateData[name]
}

func (a *Artifact) Destroy() error {
	return nil
}
