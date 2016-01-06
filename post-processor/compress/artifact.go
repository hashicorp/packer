package compress

import (
	"fmt"
	"os"
)

const BuilderId = "packer.post-processor.compress"

type Artifact struct {
	Path     string
	Provider string
	files    []string
}

func NewArtifact(provider, path string) *Artifact {
	return &Artifact{
		Path:     path,
		Provider: provider,
	}
}

func (a *Artifact) BuilderId() string {
	return BuilderId
}

func (*Artifact) Id() string {
	return ""
}

func (a *Artifact) Files() []string {
	return []string{a.Path}
}

func (a *Artifact) String() string {
	return fmt.Sprintf("'%s' compressing: %s", a.Provider, a.Path)
}

func (*Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	return os.Remove(a.Path)
}
