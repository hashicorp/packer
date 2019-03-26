package yandex

import (
	"fmt"
	"log"
)

//revive:disable:var-naming

// Artifact represents a image as the result of a Packer build.
type Artifact struct {
	image  *Image
	driver Driver
	config *Config
}

// BuilderID returns the builder Id.
//revive:disable:var-naming
func (*Artifact) BuilderId() string {
	return BuilderID
}

// Destroy destroys the image represented by the artifact.
func (a *Artifact) Destroy() error {
	log.Printf("Destroying image: %s", a.image.Name)
	errCh := a.driver.DeleteImage(a.image.Name)
	return errCh
}

// Files returns the files represented by the artifact.
func (*Artifact) Files() []string {
	return nil
}

// Id returns the image name.
//revive:disable:var-naming
func (a *Artifact) Id() string {
	return a.image.Name
}

// String returns the string representation of the artifact.
func (a *Artifact) String() string {
	return fmt.Sprintf("A disk image was created: %v (id: %v)", a.image.Name, a.image.ID)
}

func (a *Artifact) State(name string) interface{} {
	switch name {
	case "ImageID":
		return a.image.ID
	case "ImageName":
		return a.image.Name
	case "ImageSizeGb":
		return a.image.SizeGb
	case "FolderID":
		return a.config.FolderID
	case "BuildZone":
		return a.config.Zone
	}
	return nil
}
