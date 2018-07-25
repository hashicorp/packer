package classic

import (
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
)

// Artifact is an artifact implementation that contains Image List
// and Machine Image info.
type Artifact struct {
	MachineImageName string
	MachineImageFile string
	ImageListVersion int
	driver           *compute.ComputeClient
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

func (a *Artifact) Id() string {
	return a.MachineImageName
}

func (a *Artifact) String() string {
	return fmt.Sprintf("An image list entry was created: \n"+
		"Name: %s\n"+
		"File: %s\n"+
		"Version: %d",
		a.MachineImageName, a.MachineImageFile, a.ImageListVersion)
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

// Destroy deletes the custom image associated with the artifact.
func (a *Artifact) Destroy() error {
	return nil
}
