package triton

import (
	"fmt"
	"log"
)

// Artifact is an artifact implementation that contains built Triton images.
type Artifact struct {
	// ImageID is the image ID of the artifact
	ImageID string

	// BuilderIDValue is the unique ID for the builder that created this Image
	BuilderIDValue string

	// SDC connection for cleanup etc
	Driver Driver
}

func (a *Artifact) BuilderId() string {
	return a.BuilderIDValue
}

func (*Artifact) Files() []string {
	return nil
}

func (a *Artifact) Id() string {
	return a.ImageID
}

func (a *Artifact) String() string {
	return fmt.Sprintf("Image was created: %s", a.ImageID)
}

func (a *Artifact) State(name string) interface{} {
	//TODO(jen20): Figure out how to make this work with Atlas
	return nil
}

func (a *Artifact) Destroy() error {
	log.Printf("Deleting image ID (%s)", a.ImageID)
	err := a.Driver.DeleteImage(a.ImageID)
	if err != nil {
		return err
	}

	return nil
}
