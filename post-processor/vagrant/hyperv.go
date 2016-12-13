package vagrant

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
	"os"
	"path/filepath"
	"strings"
)

type HypervProvider struct{}

func (p *HypervProvider) KeepInputArtifact() bool {
	return false
}

func (p *HypervProvider) Process(ui packer.Ui, artifact packer.Artifact, dir string) (vagrantfile string, metadata map[string]interface{}, err error) {
	// Create the metadata
	metadata = map[string]interface{}{"provider": "hyperv"}

	// ui.Message(fmt.Sprintf("artifacts all: %+v", artifact))
	var outputDir string

	// Vargant requires specific dir structure for hyperv
	// hyperv builder creates the structure in the output dir
	// we have to keep the structure in a temp dir
	// hack little bit but string in artifact usually have output dir
	artifactString := artifact.String()
	d := strings.Split(artifactString, ": ")
	outputDir = d[1]
	// ui.Message(fmt.Sprintf("artifact dir from string: %s", outputDir))

	// Copy all of the original contents into the temporary directory
	for _, path := range artifact.Files() {
		ui.Message(fmt.Sprintf("Copying: %s", path))

		var rel string

		rel, err = filepath.Rel(outputDir, filepath.Dir(path))
		// ui.Message(fmt.Sprintf("rel is: %s", rel))

		if err != nil {
			ui.Message(fmt.Sprintf("err in: %s", rel))
			return
		}

		dstDir := filepath.Join(dir, rel)
		// ui.Message(fmt.Sprintf("dstdir is: %s", dstDir))
		if _, err = os.Stat(dstDir); err != nil {
			if err = os.MkdirAll(dstDir, 0755); err != nil {
				ui.Message(fmt.Sprintf("err in creating: %s", dstDir))
				return
			}
		}

		dstPath := filepath.Join(dstDir, filepath.Base(path))

		if err = CopyContents(dstPath, path); err != nil {
			ui.Message(fmt.Sprintf("err in copying: %s to %s", path, dstPath))
			return
		}
		ui.Message(fmt.Sprintf("Copyed %s to %s", path, dstPath))
	}

	return
}
