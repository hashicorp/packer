package yandeximport

import (
	"fmt"
)

const BuilderId = "packer.post-processor.yandex-import"

type Artifact struct {
	imageID     string
	imageSource string
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Id() string {
	return a.imageSource
}

func (a *Artifact) Files() []string {
	return nil
}

func (a *Artifact) String() string {
	return fmt.Sprintf("Create image %v from source URL/image %v", a.imageID, a.imageSource)
}

func (*Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	return nil
}
