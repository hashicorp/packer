package vagrant

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
	"path/filepath"
)

type LxcProvider struct{}

func (p *LxcProvider) KeepInputArtifact() bool {
	return false
}

func (p *LxcProvider) Process(ui packer.Ui, artifact packer.Artifact, dir string) (vagrantfile string, metadata map[string]interface{}, err error) {
	// Create the metadata
	metadata = map[string]interface{}{"provider": "lxc"}

	// Copy all of the original contents into the temporary directory
	for _, path := range artifact.Files() {
		ui.Message(fmt.Sprintf("Copying: %s", path))

		dstPath := filepath.Join(dir, filepath.Base(path))
		if err = CopyContents(dstPath, path); err != nil {
			return
		}
	}

	return
}
