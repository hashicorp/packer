package vagrant

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
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
	sizeString := artifact.State("diskSize").(string)
	// sizes defined at https://github.com/qemu/qemu/blob/dd5b0f95490883cd8bc7d070db8de70d5c979cbc/qemu-img.c#L141
	// All are multiples of 1024, not 1000
	sizeString = strings.ToLower(sizeString)
	re := regexp.MustCompile(`^[\d]+(b|k|m|g|t){0,1}$`)
	matched := re.MatchString(sizeString)
	if !matched {
		return "", nil, fmt.Errorf("Malformed diskSize %v", sizeString)
	}
	value := sizeString[:len(sizeString)-1]
	suffix := sizeString[len(sizeString)-1:]
	origSize, err := strconv.ParseUint(value, 10, 64)
	var sizeGb uint64
	if err != nil {
		return "", nil, err
	}
	switch suffix {
	case "e":
		sizeGb = origSize * 1024 * 1024 * 1024
	case "p":
		sizeGb = origSize * 1024 * 1024
	case "t":
		sizeGb = origSize * 1024
	case "g":
		sizeGb = origSize
	case "m":
		sizeGb = origSize / 1024 // In MB, want GB
		if origSize%1024 > 0 {
			// Make sure we don't make the size smaller
			sizeGb++
		}
	case "k":
	case "b":
		return "", nil, fmt.Errorf("Disk size too small. Must be > 1MB")
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
		"virtual_size": sizeGb,
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
