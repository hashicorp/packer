package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer/packer"
)

// BuilderId for the local artifacts
const BuilderId = "mitchellh.vmware"

// Artifact is the result of running the VMware builder, namely a set
// of files associated with the resulting machine.
type localArtifact struct {
	id  string
	dir string
	f   []string
}

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

	return &localArtifact{
		id:  id,
		dir: dir,
		f:   files,
	}, nil
}

func (a *localArtifact) BuilderId() string {
	return BuilderId
}

func (a *localArtifact) Files() []string {
	return a.f
}

func (a *localArtifact) Id() string {
	return a.id
}

func (a *localArtifact) String() string {
	return fmt.Sprintf("VM files in directory: %s", a.dir)
}

func (a *localArtifact) State(name string) interface{} {
	return nil
}

func (a *localArtifact) Destroy() error {
	return os.RemoveAll(a.dir)
}
