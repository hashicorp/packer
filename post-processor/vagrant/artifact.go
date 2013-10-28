package vagrant

import (
	"fmt"
	"os"
)

const TypeId = "mitchellh.post-processor.vagrant"

type Artifact struct {
	Path     string
	Provider string
}

func NewArtifact(provider, path string) *Artifact {
	return &Artifact{
		Path:     path,
		Provider: provider,
	}
}

func (*Artifact) TypeId() string {
	return TypeId
}

func (a *Artifact) Files() []string {
	return []string{a.Path}
}

func (a *Artifact) Id() string {
	return ""
}

func (a *Artifact) String() string {
	return fmt.Sprintf("'%s' provider box: %s", a.Provider, a.Path)
}

func (a *Artifact) Destroy() error {
	return os.Remove(a.Path)
}
