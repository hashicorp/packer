package common

import (
	"fmt"
	"os"
	"path/filepath"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// This is the common builder ID to all of these artifacts.
const BuilderId = "mitchellh.virtualbox"

// Artifact is the result of running the VirtualBox builder, namely a set
// of files associated with the resulting machine.
type artifact struct {
	dir string
	f   []string

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

// NewArtifact returns a VirtualBox artifact containing the files
// in the given directory.
func NewArtifact(dir string, generatedData map[string]interface{}) (packersdk.Artifact, error) {
	files := make([]string, 0, 5)
	visit := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}

		return err
	}

	if err := filepath.Walk(dir, visit); err != nil {
		return nil, err
	}

	return &artifact{
		dir:       dir,
		f:         files,
		StateData: generatedData,
	}, nil
}

func (*artifact) BuilderId() string {
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
	return a.StateData[name]
}

func (a *artifact) Destroy() error {
	return os.RemoveAll(a.dir)
}
