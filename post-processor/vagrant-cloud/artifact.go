package vagrantcloud

import (
	"fmt"
)

const BuilderId = "pearkes.post-processor.vagrant-cloud"

type Artifact struct {
	Tag      string
	Provider string
}

func NewArtifact(provider, tag string) *Artifact {
	return &Artifact{
		Tag:      tag,
		Provider: provider,
	}
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return nil
}

func (a *Artifact) Id() string {
	return ""
}

func (a *Artifact) String() string {
	return fmt.Sprintf("'%s': %s", a.Provider, a.Tag)
}

func (a *Artifact) Destroy() error {
	return nil
}
