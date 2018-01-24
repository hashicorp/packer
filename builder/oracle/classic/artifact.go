package classic

import (
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
)

// Artifact is an artifact implementation that contains a Snapshot.
type Artifact struct {
	Snapshot *compute.Snapshot
	driver   *compute.ComputeClient
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
	return a.Snapshot.Name
}

func (a *Artifact) String() string {
	return fmt.Sprintf("A Snapshot was created: \n"+
		"Name: %s\n"+
		"Instance: %s\n"+
		"MachineImage: %s\n"+
		"URI: %s",
		a.Snapshot.Name, a.Snapshot.Instance, a.Snapshot.MachineImage, a.Snapshot.URI)
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

// Destroy deletes the custom image associated with the artifact.
func (a *Artifact) Destroy() error {
	client := a.driver.Snapshots()
	mic := a.driver.MachineImages()
	input := &compute.DeleteSnapshotInput{
		Snapshot:     a.Snapshot.Name,
		MachineImage: a.Snapshot.MachineImage,
	}
	return client.DeleteSnapshot(mic, input)
}
