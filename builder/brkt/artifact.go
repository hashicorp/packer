package brkt

import (
	"fmt"

	"github.com/brkt/brkt-sdk-go/brkt"
)

type Artifact struct {
	ImageId        string
	ImageName      string
	BuilderIdValue string
	ApiClient      *brkt.ApiClient
}

func (a Artifact) BuilderId() string {
	return a.BuilderIdValue
}

func (a Artifact) Files() []string {
	// We have no files
	return nil
}

func (a Artifact) Id() string {
	return a.ImageId
}

func (a Artifact) String() string {
	return fmt.Sprintf("Images added to Image Catalog: \n\n\t%s\t%s", a.ImageId, a.ImageName)
}

func (a Artifact) State(name string) interface{} {
	// Not clear what state that post-processors could be querying us for
	return nil
}

func (a Artifact) Destroy() error {
	id := &brkt.ImageDefinition{
		Data: &brkt.ImageDefinitionData{
			Id: a.ImageId,
		},
		ApiClient: a.ApiClient,
	}

	return id.Delete()
}
