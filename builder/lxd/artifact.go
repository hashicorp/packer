package lxd

import (
	"fmt"
	"os"
)

type Artifact struct {
	dir string
	f   []string
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return a.f
}

func (*Artifact) Id() string {
	return "Container"
}

func (a *Artifact) String() string {
	return fmt.Sprintf("Container files in directory: %s", a.dir)
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	return os.RemoveAll(a.dir)
}
