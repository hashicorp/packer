package lxd

import (
	"fmt"
)

type Artifact struct {
	id string
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
	return nil
}

func (a *Artifact) Destroy() error {
	//return os.RemoveAll(a.dir)
	//TODO
	return nil
}
