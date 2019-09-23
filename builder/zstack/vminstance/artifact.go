package vminstance

import (
	"fmt"
	"strings"

	"github.com/hashicorp/packer/builder/zstack/zstacktype"
)

// Artifact represents a ZStack image as the result of a Packer build.
type Artifact struct {
	// BuilderId is the unique ID for the builder that created this alicloud image
	builderIdValue string
	driver         Driver
	config         Config
	images         []*zstacktype.Image
	exportPath     []string
}

// BuilderId returns the builder Id.
func (a *Artifact) BuilderId() string {
	return a.builderIdValue
}

// Destroy destroys the ZStack image represented by the artifact.
func (a *Artifact) Destroy() error {
	return nil
}

// Files returns the files represented by the artifact.
func (a *Artifact) Files() []string {
	if len(a.exportPath) > 0 {
		return a.exportPath
	} else {
		return nil
	}
}

// Id returns the ZStack image uuid.
func (a *Artifact) Id() string {
	for _, v := range a.images {
		if strings.HasSuffix(v.Name, "-Root") {
			return v.Uuid
		}
	}
	return a.images[0].Uuid
}

// String returns the string representation of the artifact.
func (a *Artifact) String() string {
	return fmt.Sprintf("A zstack image was created")
}

func (a *Artifact) State(name string) interface{} {
	return fmt.Sprintf("State: name - %s", name)
}
