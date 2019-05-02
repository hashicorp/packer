package yandex

import (
	"fmt"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

type Artifact struct {
	config *Config
	driver Driver
	image  *compute.Image
}

//revive:disable:var-naming
func (*Artifact) BuilderId() string {
	return BuilderID
}

func (a *Artifact) Id() string {
	return a.image.Id
}

func (*Artifact) Files() []string {
	return nil
}

//revive:enable:var-naming
func (a *Artifact) String() string {
	return fmt.Sprintf("A disk image was created: %v (id: %v) with family name %v", a.image.Name, a.image.Id, a.image.Family)
}

func (a *Artifact) State(name string) interface{} {
	switch name {
	case "ImageID":
		return a.image.Id
	case "FolderID":
		return a.image.FolderId
	}
	return nil

}

func (a *Artifact) Destroy() error {
	return a.driver.DeleteImage(a.image.Id)
}
