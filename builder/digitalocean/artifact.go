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

	// The name of the region
	regionName string

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
	// mimicing the aws builder
	return fmt.Sprintf("%s:%s", a.regionName, a.snapshotName)
}

func (a *Artifact) String() string {
	return fmt.Sprintf("A snapshot was created: '%v' in region '%v'", a.snapshotName, a.regionName)
}

func (a *Artifact) Destroy() error {
	log.Printf("Destroying image: %d (%s)", a.snapshotId, a.snapshotName)
	return a.client.DestroyImage(a.snapshotId)
}
