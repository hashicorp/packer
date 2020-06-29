package lxd

import (
	"fmt"
)

type Artifact struct {
	id string

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
	client lxdClient
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return nil
}

func (a *Artifact) Id() string {
	return a.id
}

func (a *Artifact) String() string {
	return fmt.Sprintf("image: %s", a.id)
}

func (a *Artifact) State(name string) interface{} {
	return a.StateData[name]
}

func (a *Artifact) Destroy() error {
	return a.client.DeleteImage(a.id)
}
