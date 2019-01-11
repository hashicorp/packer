package vagrant

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/packer/packer"
)

// This is the common builder ID to all of these artifacts.
const BuilderId = "vagrant"

// Artifact is the result of running the vagrant builder, namely a set
// of files associated with the resulting machine.
type artifact struct {
	OutputDir string
	BoxName   string
}

// NewArtifact returns a vagrant artifact containing the .box file
func NewArtifact(dir string) (packer.Artifact, error) {
	return &artifact{
		OutputDir: dir,
		BoxName:   "package.box",
	}, nil
}

func (*artifact) BuilderId() string {
	return BuilderId
}

func (a *artifact) Files() []string {
	return []string{a.BoxName}
}

func (a *artifact) Id() string {
	return filepath.Join(a.OutputDir, a.BoxName)
}

func (a *artifact) String() string {
	return fmt.Sprintf("Vagrant box is  %s", a.Id())
}

func (a *artifact) State(name string) interface{} {
	return nil
}

func (a *artifact) Destroy() error {
	return nil
}
