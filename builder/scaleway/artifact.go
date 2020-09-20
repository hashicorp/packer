package scaleway

import (
	"fmt"
	"log"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
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

	// The name of the zone
	zoneName string

	// The client for making API calls
	client *scw.Client

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (*Artifact) Files() []string {
	// No files with Scaleway
	return nil
}

func (a *Artifact) Id() string {
	return fmt.Sprintf("%s:%s", a.zoneName, a.imageID)
}

func (a *Artifact) String() string {
	return fmt.Sprintf("An image was created: '%v' (ID: %v) in zone '%v' based on snapshot '%v' (ID: %v)",
		a.imageName, a.imageID, a.zoneName, a.snapshotName, a.snapshotID)
}

func (a *Artifact) State(name string) interface{} {
	return a.StateData[name]
}

func (a *Artifact) Destroy() error {
	log.Printf("Destroying image: %s (%s)", a.imageID, a.imageName)
	instanceAPI := instance.NewAPI(a.client)

	err := instanceAPI.DeleteImage(&instance.DeleteImageRequest{
		ImageID: a.imageID,
	})
	if err != nil {
		return err
	}
	log.Printf("Destroying snapshot: %s (%s)", a.snapshotID, a.snapshotName)
	err = instanceAPI.DeleteSnapshot(&instance.DeleteSnapshotRequest{
		SnapshotID: a.snapshotID,
	})
	if err != nil {
		return err
	}
	return nil
}
