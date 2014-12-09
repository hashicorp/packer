package atlas

import (
	"fmt"
)

const BuilderId = "packer.post-processor.atlas"

type Artifact struct {
	Name    string
	Type    string
	Version int
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return nil
}

func (a *Artifact) Id() string {
	return fmt.Sprintf("%s/%s/%d", a.Name, a.Type, a.Version)
}

func (a *Artifact) String() string {
	return fmt.Sprintf("%s/%s (v%d)", a.Name, a.Type, a.Version)
}

func (*Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	return nil
}
