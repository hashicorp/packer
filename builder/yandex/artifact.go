package yandex

import (
	"fmt"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

type Artifact struct {
	config *Config
	driver Driver
	Image  *compute.Image

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

//revive:disable:var-naming
func (*Artifact) BuilderId() string {
	return BuilderID
}

func (a *Artifact) Id() string {
	return a.Image.Id
}

func (*Artifact) Files() []string {
	return []string{""}
}

//revive:enable:var-naming
func (a *Artifact) String() string {
	return fmt.Sprintf("A disk image was created: %v (id: %v) with family name %v", a.Image.Name, a.Image.Id, a.Image.Family)
}

func (a *Artifact) State(name string) interface{} {
	if _, ok := a.StateData[name]; ok {
		return a.StateData[name]
	}

	switch name {
	case "ImageID":
		return a.Image.Id
	case "FolderID":
		return a.Image.FolderId
	}
	return nil

}

func (a *Artifact) Destroy() error {
	return a.driver.DeleteImage(a.Image.Id)
}
