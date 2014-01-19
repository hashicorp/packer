package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/packer/packer"
)

// BuilderId for the local artifacts
const BuilderId = "mitchellh.vmware"

// Artifact is the result of running the VMware builder, namely a set
// of files associated with the resulting machine.
type localArtifact struct {
	dir string
	f   []string
}

// NewLocalArtifact returns a VMware artifact containing the files
// in the given directory.
func NewLocalArtifact(dir string) (packer.Artifact, error) {
	files := make([]string, 0, 5)
	visit := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}

		return err
	}

	if err := filepath.Walk(dir, visit); err != nil {
		return nil, err
	}

	return &localArtifact{
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

func (*localArtifact) Id() string {
	return "VM"
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
