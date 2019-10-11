package scaleway

import (
	"fmt"
	"log"

	"github.com/scaleway/scaleway-cli/pkg/api"
)

type Artifact struct {
	// The name of the image
	imageName string

	// The ID of the image
	imageID string

	// The name of the snapshot
	snapshotName string

	// The ID of the snapshot
	snapshotID string

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
	return fmt.Sprintf("%s:%s", a.regionName, a.imageID)
}

func (a *Artifact) String() string {
	return fmt.Sprintf("An image was created: '%v' (ID: %v) in region '%v' based on snapshot '%v' (ID: %v)",
		a.imageName, a.imageID, a.regionName, a.snapshotName, a.snapshotID)
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	log.Printf("Destroying image: %s (%s)", a.imageID, a.imageName)
	if err := a.client.DeleteImage(a.imageID); err != nil {
		return err
	}
	log.Printf("Destroying snapshot: %s (%s)", a.snapshotID, a.snapshotName)
	if err := a.client.DeleteSnapshot(a.snapshotID); err != nil {
		return err
	}
	return nil
}
