package common

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer/packer"
)

// BuilderId for the local artifacts
const BuilderId = "mitchellh.vmware"
const BuilderIdESX = "mitchellh.vmware-esx"

// Artifact is the result of running the VMware builder, namely a set
// of files associated with the resulting machine.
type artifact struct {
	builderId string
	id        string
	dir       string
	f         []string
	config    map[string]string
}

// NewLocalArtifact returns a VMware artifact containing the files
// in the given directory.
// NewLocalArtifact returns a VMware artifact containing the files
// in the given directory.
func NewLocalArtifact(id string, dir string) (packer.Artifact, error) {
	files := make([]string, 0, 5)
	visit := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	}

	if err := filepath.Walk(dir, visit); err != nil {
		return nil, err
	}

	return &artifact{
		builderId: id,
		dir:       dir,
		f:         files,
	}, nil
}

func NewArtifact(dir OutputDir, files []string, esxi bool) (packer.Artifact, err) {
	builderID := BuilderId
	if esxi {
		builderID = BuilderIdESX
	}

	return &artifact{
		builderId: builderID,
		dir:       dir.String(),
		f:         files,
	}, nil
}

func (a *artifact) BuilderId() string {
	return BuilderId
}

func (a *artifact) Files() []string {
	return a.f
}

func (*artifact) Id() string {
	return a.id
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
