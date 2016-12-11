package vagrant

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
	"path/filepath"
	"os"
)

type HypervProvider struct{}

func (p *HypervProvider) KeepInputArtifact() bool {
	return false
}

func (p *HypervProvider) Process(ui packer.Ui, artifact packer.Artifact, dir string) (vagrantfile string, metadata map[string]interface{}, err error) {
	// Create the metadata
	metadata = map[string]interface{}{"provider": "hyperv"}

	var outputDir string

	// Copy all of the original contents into the temporary directory
	for i, path := range artifact.Files() {

		// Vargant requires specific dir structure for hyperv
		// hyperv builder creates the structure in the output dir
		// we have to keep the structure in a temp dir

		// hyperv artifact put the output dir path as the first elem
		if i == 0 {
			outputDir = path
			continue
		}

		ui.Message(fmt.Sprintf("Copying: %s", path))

		var rel string

		rel, err = filepath.Rel(outputDir, filepath.Dir(path))

		if err != nil {
			return
		}

		dstDir := filepath.Join(dir, rel)

		if _, err = os.Stat(dstDir); err != nil {
			if err = os.MkdirAll(dstDir, 0755); err != nil {
				return
			}
		}

		dstPath := filepath.Join(dstDir, filepath.Base(path))
		if err = CopyContents(dstPath, path); err != nil {
			return
		}
	}

	return
}
