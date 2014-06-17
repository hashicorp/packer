package vagrant

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/mitchellh/packer/packer"
)

// These are the extensions of files and directories that are unnecessary for the function
// of a Parallels virtual machine.
var UnnecessaryFilesPatterns = []string{"\\.log$", "\\.backup$", "\\.Backup$", "\\.app/"}

type ParallelsProvider struct{}

func (p *ParallelsProvider) KeepInputArtifact() bool {
	return false
}

func (p *ParallelsProvider) Process(ui packer.Ui, artifact packer.Artifact, dir string) (vagrantfile string, metadata map[string]interface{}, err error) {
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

	// Create the Vagrantfile from the template
	vagrantfile = fmt.Sprintf(parallelsVagrantfile)

	return
}

var parallelsVagrantfile = `
Vagrant.configure("2") do |config|
end
`
