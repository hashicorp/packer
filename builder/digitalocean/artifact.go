package digitalocean

import (
	"fmt"
	"log"
)

type Artifact struct {
	// The name of the snapshot
	snapshotName string

	// The ID of the image
	snapshotId uint

	// The client for making API calls
	client *DigitalOceanClient
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (*Artifact) Files() []string {
	// No files with DigitalOcean
	return nil
}

func (a *Artifact) Id() string {
	return a.snapshotName
}

func (a *Artifact) String() string {
	return fmt.Sprintf("A snapshot was created: %v", a.snapshotName)
}

func (a *Artifact) Destroy() error {
	log.Printf("Destroying image: %d", a.snapshotId)
	return a.client.DestroyImage(a.snapshotId)
}
