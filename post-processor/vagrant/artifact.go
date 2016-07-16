package vagrant

import (
	"fmt"
	"os"
)

const BuilderId = "mitchellh.post-processor.vagrant"

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

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return []string{a.Path}
}

func (a *Artifact) Id() string {
	return a.Provider
}

func (a *Artifact) String() string {
	return fmt.Sprintf("'%s' provider box: %s", a.Provider, a.Path)
}

func (a *Artifact) State(name string) interface{} {
	switch name {
	case "atlas.artifact.metadata":
		return a.stateAtlasMetadata()
	default:
		return nil
	}
}

func (a *Artifact) Destroy() error {
	return os.Remove(a.Path)
}

func (a *Artifact) stateAtlasMetadata() map[string]string {
	return map[string]string{"provider": a.vagrantProvider()}
}

func (a *Artifact) vagrantProvider() string {
	switch a.Provider {
	case "digitalocean":
		return "digital_ocean"
	case "vmware":
		return "vmware_desktop"
	default:
		return a.Provider
	}
}
