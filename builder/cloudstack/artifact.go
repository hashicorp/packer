package cloudstack

import (
	"fmt"
	"log"
	"strings"

	"github.com/xanzy/go-cloudstack/cloudstack"
)

// Artifact represents a CloudStack template as the result of a Packer build.
type Artifact struct {
	client   *cloudstack.CloudStackClient
	config   *Config
	template *cloudstack.CreateTemplateResponse
}

// BuilderId returns the builder ID.
func (a *Artifact) BuilderId() string {
	return BuilderId
}

// Destroy the CloudStack template represented by the artifact.
func (a *Artifact) Destroy() error {
	// Create a new parameter struct.
	p := a.client.Template.NewDeleteTemplateParams(a.template.Id)

	// Destroy the template.
	log.Printf("Destroying template: %s", a.template.Name)
	_, err := a.client.Template.DeleteTemplate(p)
	if err != nil {
		// This is a very poor way to be told the ID does no longer exist :(
		if strings.Contains(err.Error(), fmt.Sprintf(
			"Invalid parameter id value=%s due to incorrect long value format, "+
				"or entity does not exist", a.template.Id)) {
			return nil
		}

		return fmt.Errorf("Error destroying template %s: %s", a.template.Name, err)
	}

	return nil
}

// Files returns the files represented by the artifact.
func (a *Artifact) Files() []string {
	// We have no files.
	return nil
}

// Id returns CloudStack template ID.
func (a *Artifact) Id() string {
	return a.template.Id
}

// String returns the string representation of the artifact.
func (a *Artifact) String() string {
	return fmt.Sprintf("A template was created: %s", a.template.Name)
}

// State returns specific details from the artifact.
func (a *Artifact) State(name string) interface{} {
	return nil
}
