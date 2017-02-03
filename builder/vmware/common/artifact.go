package common

import (
	"fmt"
	"os"

	"github.com/mitchellh/packer/packer"
)

// BuilderId for the local artifacts
const BuilderId = "mitchellh.vmware"
const BuilderIdESX = "mitchellh.vmware-esx"

// Artifact is the result of running the VMware builder, namely a set
// of files associated with the resulting machine.
type artifact struct {
	builderId string
	dir       string
	f         []string
}

// NewLocalArtifact returns a VMware artifact containing the files
// in the given directory.
func NewArtifact(dir OutputDir, files []string, esxi bool) packer.Artifact {
	builderID := BuilderId
	if esxi {
		builderID = BuilderIdESX
	}
	return &artifact{
		builderId: builderID,
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

func (*artifact) Id() string {
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
