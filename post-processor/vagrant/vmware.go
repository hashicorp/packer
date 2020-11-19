package vagrant

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type VMwareProvider struct{}

func (p *VMwareProvider) KeepInputArtifact() bool {
	return false
}

func (p *VMwareProvider) Process(ui packersdk.Ui, artifact packer.Artifact, dir string) (vagrantfile string, metadata map[string]interface{}, err error) {
	// Create the metadata
	metadata = map[string]interface{}{"provider": "vmware_desktop"}

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
