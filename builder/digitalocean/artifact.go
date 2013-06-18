package digitalocean

import (
	"errors"
	"fmt"
)

type Artifact struct {
	// The name of the snapshot
	snapshotName string
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
	return errors.New("not implemented yet")
}
