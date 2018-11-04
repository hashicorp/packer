package vagrant

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer/packer"
)

type LibVirtProvider struct{}

func (p *LibVirtProvider) KeepInputArtifact() bool {
	return false
}
func (p *LibVirtProvider) Process(ui packer.Ui, artifact packer.Artifact, dir string) (vagrantfile string, metadata map[string]interface{}, err error) {
	diskName := artifact.State("diskName").(string)

	// Copy the disk image into the temporary directory (as box.img)
	for _, path := range artifact.Files() {
		if strings.HasSuffix(path, "/"+diskName) {
			ui.Message(fmt.Sprintf("Copying from artifact: %s", path))
			dstPath := filepath.Join(dir, "box.img")
			if err = CopyContents(dstPath, path); err != nil {
				return
			}
		}
	}

	format := artifact.State("diskType").(string)
	origSize := artifact.State("diskSize").(uint64)
	size := origSize / 1024 // In MB, want GB
	if origSize%1024 > 0 {
		// Make sure we don't make the size smaller
		size++
	}
	domainType := artifact.State("domainType").(string)

	// Convert domain type to libvirt driver
	var driver string
	switch domainType {
	case "none", "tcg", "hvf":
		driver = "qemu"
	case "kvm":
		driver = domainType
	default:
		return "", nil, fmt.Errorf("Unknown libvirt domain type: %s", domainType)
	}

	// Create the metadata
	metadata = map[string]interface{}{
		"provider":     "libvirt",
		"format":       format,
		"virtual_size": size,
	}

	vagrantfile = fmt.Sprintf(libvirtVagrantfile, driver)
	return
}

var libvirtVagrantfile = `
Vagrant.configure("2") do |config|
  config.vm.provider :libvirt do |libvirt|
    libvirt.driver = "%s"
  end
end
`
