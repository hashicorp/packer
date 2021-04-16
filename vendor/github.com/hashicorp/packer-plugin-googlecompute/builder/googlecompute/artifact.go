package googlecompute

import (
	"fmt"
	"log"
)

// Artifact represents a GCE image as the result of a Packer build.
type Artifact struct {
	image  *Image
	driver Driver
	config *Config
	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

// BuilderId returns the builder Id.
func (*Artifact) BuilderId() string {
	return BuilderId
}

// Destroy destroys the GCE image represented by the artifact.
func (a *Artifact) Destroy() error {
	log.Printf("Destroying image: %s", a.image.Name)
	errCh := a.driver.DeleteImage(a.image.Name)
	return <-errCh
}

// Files returns the files represented by the artifact.
func (*Artifact) Files() []string {
	return nil
}

// Id returns the GCE image name.
func (a *Artifact) Id() string {
	return a.image.Name
}

// String returns the string representation of the artifact.
func (a *Artifact) String() string {
	return fmt.Sprintf("A disk image was created: %v", a.image.Name)
}

func (a *Artifact) State(name string) interface{} {
	if _, ok := a.StateData[name]; ok {
		return a.StateData[name]
	}

	switch name {
	case "ImageName":
		return a.image.Name
	case "ImageSizeGb":
		return a.image.SizeGb
	case "AccountFilePath":
		return a.config.AccountFile
	case "ProjectId":
		return a.config.ProjectId
	case "BuildZone":
		return a.config.Zone
	}
	return nil
}
