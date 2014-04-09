package vagrant

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/going/toolkit/xmlpath"
	"github.com/mitchellh/packer/packer"
)

// These are the extensions of files that are unnecessary for the function
// of a Parallels virtual machine.
var UnnecessaryFileExtensions = []string{".log", ".backup", ".Backup"}

type ParallelsProvider struct{}

func (p *ParallelsProvider) KeepInputArtifact() bool {
	return false
}

func (p *ParallelsProvider) Process(ui packer.Ui, artifact packer.Artifact, dir string) (vagrantfile string, metadata map[string]interface{}, err error) {
	// Create the metadata
	metadata = map[string]interface{}{"provider": "parallels"}
	var configPath string

	// Copy all of the original contents into the temporary directory
	for _, path := range artifact.Files() {
		// If the file isn't critical to the function of the
		// virtual machine, we get rid of it.
		// It's done by the builder, but we need one more time
		// because unregistering a vm creates config.pvs.backup again.
		unnecessary := false
		ext := filepath.Ext(path)
		for _, unnecessaryExt := range UnnecessaryFileExtensions {
			if unnecessaryExt == ext {
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
		if strings.HasSuffix(dstPath, "/config.pvs") {
			configPath = dstPath
		}
	}

	// Create the Vagrantfile from the template
	var baseMacAddress string
	baseMacAddress, err = findBaseMacAddress(configPath)
	if err != nil {
		ui.Message(fmt.Sprintf("Problem determining Vagarant Box MAC address: %s", err))
	}

	vagrantfile = fmt.Sprintf(parallelsVagrantfile, baseMacAddress)

	return
}

func findBaseMacAddress(path string) (string, error) {
	xpath := "/ParallelsVirtualMachine/Hardware/NetworkAdapter[@id='0']/MAC"
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	xpathComp := xmlpath.MustCompile(xpath)
	root, err := xmlpath.Parse(file)
	if err != nil {
		return "", err
	}
	value, _ := xpathComp.String(root)
	return value, nil
}

var parallelsVagrantfile = `
Vagrant.configure("2") do |config|
  config.vm.base_mac = "%s"
end
`
