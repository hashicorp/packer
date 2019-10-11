package linode

import (
	"context"
	"fmt"
	"log"

	"github.com/linode/linodego"
)

type Artifact struct {
	ImageID    string
	ImageLabel string

	Driver *linodego.Client
}

func (a Artifact) BuilderId() string { return BuilderID }
func (a Artifact) Files() []string   { return nil }
func (a Artifact) Id() string        { return a.ImageID }

func (a Artifact) String() string {
	return fmt.Sprintf("Linode image: %s (%s)", a.ImageLabel, a.ImageID)
}

func (a Artifact) State(name string) interface{} { return nil }

func (a Artifact) Destroy() error {
	log.Printf("Destroying image: %s (%s)", a.ImageID, a.ImageLabel)
	err := a.Driver.DeleteImage(context.TODO(), a.ImageID)
	return err
}
