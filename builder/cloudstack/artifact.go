package cloudstack

import (
	"fmt"
	"github.com/mindjiver/gopherstack"
	"log"
)

type Artifact struct {
	// The name of the template
	templateName string

	// The ID of the image
	templateId string

	// The client for making API calls
	client *gopherstack.CloudStackClient
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (*Artifact) Files() []string {
	// No local files created with Cloudstack.
	return nil
}

func (a *Artifact) Id() string {
	return a.templateName
}

func (a *Artifact) String() string {
	return fmt.Sprintf("A template was created: %v", a.templateName)
}

func (a *Artifact) Destroy() error {
	log.Printf("Delete template: %s", a.templateId)
	_, err := a.client.DeleteTemplate(a.templateId)
	return err
}
