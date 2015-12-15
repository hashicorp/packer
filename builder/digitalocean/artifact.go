package digitalocean

import (
	"fmt"
	"log"
	"strconv"

	"github.com/digitalocean/godo"
)

type Artifact struct {
	// The name of the snapshot
	snapshotName string

	// The ID of the image
	snapshotId int

	// The name of the region
	regionName string

	// The client for making API calls
	client *godo.Client
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (*Artifact) Files() []string {
	// No files with DigitalOcean
	return nil
}

func (a *Artifact) Id() string {
	return fmt.Sprintf("%s:%s", a.regionName, strconv.FormatUint(uint64(a.snapshotId), 10))
}

func (a *Artifact) String() string {
	return fmt.Sprintf("A snapshot was created: '%v' (ID: %v) in region '%v'", a.snapshotName, a.snapshotId, a.regionName)
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	log.Printf("Destroying image: %d (%s)", a.snapshotId, a.snapshotName)
	_, err := a.client.Images.Delete(a.snapshotId)
	return err
}
