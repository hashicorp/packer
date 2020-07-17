package yandeximport

import (
	"fmt"
)

const BuilderId = "packer.post-processor.yandex-import"

type Artifact struct {
	imageID   string
	sourceURL string
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Id() string {
	return a.sourceURL
}

func (a *Artifact) Files() []string {
	return nil
}

func (a *Artifact) String() string {
	return fmt.Sprintf("Create image %v from URL %v", a.imageID, a.sourceURL)
}

func (*Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	return nil
}
