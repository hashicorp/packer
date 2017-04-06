package scaleway

import (
	"fmt"
	"log"

	"github.com/scaleway/scaleway-cli/pkg/api"
)

type Artifact struct {
	// The name of the snapshot
	snapshotName string

	// The ID of the snapshot
	snapshotId string

	// The name of the region
	regionName string

	// The client for making API calls
	client *api.ScalewayAPI
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (*Artifact) Files() []string {
	// No files with Scaleway
	return nil
}

func (a *Artifact) Id() string {
	return fmt.Sprintf("%s:%s", a.regionName, a.snapshotId)
}

func (a *Artifact) String() string {
	return fmt.Sprintf("A snapshot was created: '%v' (ID: %v) in region '%v'", a.snapshotName, a.snapshotId, a.regionName)
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	log.Printf("Destroying image: %s (%s)", a.snapshotId, a.snapshotName)
	err := a.client.DeleteSnapshot(a.snapshotId)
	return err
}
