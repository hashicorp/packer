package digitalocean

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/digitalocean/godo"
	"github.com/hashicorp/packer/packer"
)

// A map from volume snapshot names to volume snapshot IDs
type snapshot struct {
	// The name of the snapshot
	name string
	// The ID of the snapshot
	id string
	// The name of the regions in which the snapshot is available
	regions []string
}

type Artifact struct {
	droplet snapshot
	volumes []snapshot

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
	parts := make([]string, 0, len(a.volumes))
	for _, snapshot := range append([]snapshot{a.droplet}, a.volumes...) {
		parts = append(parts, fmt.Sprintf("%s:%s",
			strings.Join(snapshot.regions, ","), snapshot.id))
	}
	return strings.Join(parts, ",")
}

func (a *Artifact) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "A droplet snapshot was created: '%s' (ID: %s) in regions '%v'",
		a.droplet.name, a.droplet.id, strings.Join(a.droplet.regions, ","))
	if len(a.volumes) > 0 {
		fmt.Fprintf(&buf, "\nVolume snapshots were created:")
		for _, s := range a.volumes {
			fmt.Fprintf(&buf, "\n'%s' (ID: %s) in regions '%s'",
				s.name, s.id, strings.Join(s.regions, ","))
		}
	}
	return buf.String()
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	var errors []error

	log.Printf("Destroying droplet snapshot: %s (%s)", a.droplet.id, a.droplet.name)
	imageId, err := strconv.Atoi(a.droplet.id)
	if err != nil {
		errors = append(errors, err)
	} else {
		_, err := a.client.Images.Delete(context.TODO(), imageId)
		if err != nil {
			errors = append(errors, err)
		}
	}

	for _, s := range a.volumes {
		log.Printf("Destroying volume snapshot: %s (%s)", s.id, s.name)
		_, err := a.client.Storage.DeleteSnapshot(context.TODO(), s.id)
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		if len(errors) == 1 {
			return errors[0]
		} else {
			return &packer.MultiError{Errors: errors}
		}
	}

	return nil
}
