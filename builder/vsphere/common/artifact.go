package common

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer/packer"
)

const BuilderId = "packer.vsphere"

// Artifact is the result of running the VMware builder, namely a set
// of files associated with the resulting machine.
type artifact struct {
	builderId string
	dir       string
	f         []string
}

func NewArtifact(dir OutputDir, files []string) packer.Artifact {
	return &artifact{
		builderId: BuilderId,
		dir:       dir.String(),
		f:         files,
	}
}

func (a *artifact) BuilderId() string {
	return BuilderId
}

func (a *artifact) Files() []string {
	return a.f
}

func (a *artifact) Id() string {
	return "VM"
}

func (a *artifact) String() string {
	return fmt.Sprintf("VM files in directory: %s", a.dir)
}

func (a *artifact) State(name string) interface{} {
	return nil
}

func (a *artifact) Destroy() error {
	return os.RemoveAll(a.dir)
}
