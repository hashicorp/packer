package vagrant

import (
	"fmt"
	"path/filepath"
	"regexp"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// These are the extensions of files and directories that are unnecessary for the function
// of a Parallels virtual machine.
var UnnecessaryFilesPatterns = []string{"\\.log$", "\\.backup$", "\\.Backup$", "\\.app/", "/Windows Disks/"}

type ParallelsProvider struct{}

func (p *ParallelsProvider) KeepInputArtifact() bool {
	return false
}

func (p *ParallelsProvider) Process(ui packersdk.Ui, artifact packersdk.Artifact, dir string) (vagrantfile string, metadata map[string]interface{}, err error) {
	// Create the metadata
	metadata = map[string]interface{}{"provider": "parallels"}

	// Copy all of the original contents into the temporary directory
	for _, path := range artifact.Files() {
		// If the file isn't critical to the function of the
		// virtual machine, we get rid of it.
		unnecessary := false
		for _, unnecessaryPat := range UnnecessaryFilesPatterns {
			if matched, _ := regexp.MatchString(unnecessaryPat, path); matched {
				unnecessary = true
				break
			}
		}
		if unnecessary {
			continue
		}

		tmpPath := filepath.ToSlash(path)
		pathRe := regexp.MustCompile(`^(.+?)([^/]+\.pvm/.+?)$`)
		matches := pathRe.FindStringSubmatch(tmpPath)
		var pvmPath string
		if matches != nil {
			pvmPath = filepath.FromSlash(matches[2])
		} else {
			continue // Just copy a pvm
		}
		dstPath := filepath.Join(dir, pvmPath)

		ui.Message(fmt.Sprintf("Copying: %s", path))
		if err = CopyContents(dstPath, path); err != nil {
			return
		}
	}

	return
}
