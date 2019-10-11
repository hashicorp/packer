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
	Provider  string
}

// NewArtifact returns a vagrant artifact containing the .box file
func NewArtifact(provider, dir string) packer.Artifact {
	return &artifact{
		OutputDir: dir,
		BoxName:   "package.box",
		Provider:  provider,
	}
}

func (*artifact) BuilderId() string {
	return BuilderId
}

func (a *artifact) Files() []string {
	return []string{filepath.Join(a.OutputDir, a.BoxName)}
}

func (a *artifact) Id() string {
	return a.Provider
}

func (a *artifact) String() string {
	return fmt.Sprintf("Vagrant box '%s' for '%s' provider", a.BoxName, a.Provider)
}

func (a *artifact) State(name string) interface{} {
	return nil
}

func (a *artifact) Destroy() error {
	return nil
}
