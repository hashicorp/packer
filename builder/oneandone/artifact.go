package oneandone

import (
	"fmt"
)

type Artifact struct {
	snapshotId   string
	snapshotName string
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
	return fmt.Sprintf("A snapshot was created: '%v', '%v'", a.snapshotId, a.snapshotName)
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	return nil
}
