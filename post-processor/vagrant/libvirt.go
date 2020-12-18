package vagrant

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// Lowercase a ascii letter.
func lower(c byte) byte {
	return c | ('a' - 'A')
}

// Convert a string that represents a qemu disk image size to megabytes.
//
// Valid units (case-insensitive):
//
//		B (byte)        1B
//		K (kilobyte) 1024B
//		M (megabyte) 1024K
//		G (gigabyte) 1024M
//		T (terabyte) 1024G
//		P (petabyte) 1024T
//		E (exabyte)  1024P
//
// The default is M.
func sizeInMegabytes(size string) uint64 {
	unit := size[len(size)-1]

	if unit >= '0' && unit <= '9' {
		unit = 'm'
	} else {
		size = size[:len(size)-1]
	}

	value, _ := strconv.ParseUint(size, 10, 64)

	switch lower(unit) {
	case 'b':
		return value / 1024 / 1024
	case 'k':
		return value / 1024
	case 'm':
		return value
	case 'g':
		return value * 1024
	case 't':
		return value * 1024 * 1024
	case 'p':
		return value * 1024 * 1024 * 1024
	case 'e':
		return value * 1024 * 1024 * 1024 * 1024
	default:
		panic(fmt.Sprintf("Unknown size unit %c", unit))
	}
}

type LibVirtProvider struct{}

func (p *LibVirtProvider) KeepInputArtifact() bool {
	return false
}
func (p *LibVirtProvider) Process(ui packersdk.Ui, artifact packersdk.Artifact, dir string) (vagrantfile string, metadata map[string]interface{}, err error) {
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
	origSize := sizeInMegabytes(artifact.State("diskSize").(string))
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
