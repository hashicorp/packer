package vultr

import (
	"context"
	"fmt"
	"log"

	"github.com/vultr/govultr"
)

type Artifact struct {
	// The ID of the snapshot
	SnapshotID string

	// The Description of the snapshot
	Description string

	// The client for making
	client *govultr.Client
}

func (a *Artifact) BuilderId() string {
	return BuilderID
}

func (a *Artifact) Files() []string {
	return nil
}

func (a *Artifact) Id() string {
	return a.SnapshotID
}

func (a *Artifact) String() string {
	return fmt.Sprintf("Vultr Snapshot: %s (%s)", a.Description, a.SnapshotID)
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	log.Printf("Destroying Vultr Snapshot: %s (%s)", a.SnapshotID, a.Description)
	err := a.client.Snapshot.Delete(context.Background(), a.SnapshotID)
	return err
}
