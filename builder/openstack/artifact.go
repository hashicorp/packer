package openstack

import (
	"fmt"
	"github.com/rackspace/gophercloud"
	"log"
)

// Artifact is an artifact implementation that contains built images.
type Artifact struct {
	// ImageId of built image
	ImageId string

	// BuilderId is the unique ID for the builder that created this image
	BuilderIdValue string

	// OpenStack connection for performing API stuff.
	Conn gophercloud.CloudServersProvider
}

func (a *Artifact) BuilderId() string {
	return a.BuilderIdValue
}

func (*Artifact) Files() []string {
	// We have no files
	return nil
}

func (a *Artifact) Id() string {
	return a.ImageId
}

func (a *Artifact) String() string {
	return fmt.Sprintf("An image was created: %v", a.ImageId)
}

func (a *Artifact) Destroy() error {
	log.Printf("Destroying image: %d", a.ImageId)
	return a.Conn.DeleteImageById(a.ImageId)
}
