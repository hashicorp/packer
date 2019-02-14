package digitalocean

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/digitalocean/godo"
)

type Artifact struct {
	// The name of the snapshot
	SnapshotName string

	// The ID of the image
	SnapshotId int

	// The name of the region
	RegionNames []string

	// The client for making API calls
	Client *godo.Client
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (*Artifact) Files() []string {
	// No files with DigitalOcean
	return nil
}

func (a *Artifact) Id() string {
	return fmt.Sprintf("%s:%s", strings.Join(a.RegionNames[:], ","), strconv.FormatUint(uint64(a.SnapshotId), 10))
}

func (a *Artifact) String() string {
	return fmt.Sprintf("A snapshot was created: '%v' (ID: %v) in regions '%v'", a.SnapshotName, a.SnapshotId, strings.Join(a.RegionNames[:], ","))
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	log.Printf("Destroying image: %d (%s)", a.SnapshotId, a.SnapshotName)
	_, err := a.Client.Images.Delete(context.TODO(), a.SnapshotId)
	return err
}
