package yandeximport

import (
	"fmt"
)

const BuilderId = "packer.post-processor.yandex-import"

type Artifact struct {
	imageID    string
	sourceType string
	sourceID   string

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Id() string {
	return a.imageID
}

func (a *Artifact) Files() []string {
	return nil
}

func (a *Artifact) String() string {
	return fmt.Sprintf("Create image %v from source type %v with ID/URL %v", a.imageID, a.sourceType, a.sourceID)
}

func (a *Artifact) State(name string) interface{} {
	if _, ok := a.StateData[name]; ok {
		return a.StateData[name]
	}

	return nil
}

func (a *Artifact) Destroy() error {
	return nil
}
