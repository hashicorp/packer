package oci

import (
	"fmt"

	client "github.com/hashicorp/packer/builder/oracle/oci/client"
)

// Artifact is an artifact implementation that contains a built Custom Image.
type Artifact struct {
	Image  client.Image
	Region string
	driver Driver
}

// BuilderId uniquely identifies the builder.
func (a *Artifact) BuilderId() string {
	return BuilderId
}

// Files lists the files associated with an artifact. We don't have any files
// as the custom image is stored server side.
func (a *Artifact) Files() []string {
	return nil
}

// Id returns the OCID of the associated Image.
func (a *Artifact) Id() string {
	return a.Image.ID
}

func (a *Artifact) String() string {
	return fmt.Sprintf(
		"An image was created: '%v' (OCID: %v) in region '%v'",
		a.Image.DisplayName, a.Image.ID, a.Region,
	)
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

// Destroy deletes the custom image associated with the artifact.
func (a *Artifact) Destroy() error {
	return a.driver.DeleteImage(a.Image.ID)
}
