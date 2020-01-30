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

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

// NewArtifact returns a vagrant artifact containing the .box file
func NewArtifact(provider, dir string, generatedData map[string]interface{}) packer.Artifact {
	return &artifact{
		OutputDir: dir,
		BoxName:   "package.box",
		Provider:  provider,
		StateData: generatedData,
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
	return a.StateData[name]
}

func (a *artifact) Destroy() error {
	return nil
}
