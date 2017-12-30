package vagrant

import (
	"archive/tar"
	"errors"
	"fmt"
	"github.com/hashicorp/packer/packer"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

type VBoxProvider struct{}

func (p *VBoxProvider) KeepInputArtifact() bool {
	return false
}

func (p *VBoxProvider) Process(ui packer.Ui, artifact packer.Artifact, dir string) (vagrantfile string, metadata map[string]interface{}, err error) {
	// Create the metadata
	metadata = map[string]interface{}{"provider": "virtualbox"}

	// Copy all of the original contents into the temporary directory
	for _, path := range artifact.Files() {
		// We treat OVA files specially, we unpack those into the temporary
		// directory so we can get the resulting disk and OVF.
		if extension := filepath.Ext(path); extension == ".ova" {
			ui.Message(fmt.Sprintf("Unpacking OVA: %s", path))
			if err = DecompressOva(dir, path); err != nil {
				return
			}
		} else {
			ui.Message(fmt.Sprintf("Copying from artifact: %s", path))
			dstPath := filepath.Join(dir, filepath.Base(path))
			if err = CopyContents(dstPath, path); err != nil {
				return
			}
		}

	}

	// Rename the OVF file to box.ovf, as required by Vagrant
	ui.Message("Renaming the OVF to box.ovf...")
	if err = p.renameOVF(dir); err != nil {
		return
	}

	// Create the Vagrantfile from the template
	var baseMacAddress string
	baseMacAddress, err = p.findBaseMacAddress(dir)
	if err != nil {
		return
	}

	vagrantfile = fmt.Sprintf(vboxVagrantfile, baseMacAddress)
	return
}

func (p *VBoxProvider) findOvf(dir string) (string, error) {
	log.Println("Looking for OVF in artifact...")
	file_matches, err := filepath.Glob(filepath.Join(dir, "*.ovf"))
	if err != nil {
		return "", err
	}

	if len(file_matches) > 1 {
		return "", errors.New("More than one OVF file in VirtualBox artifact.")
	}

	if len(file_matches) < 1 {
		return "", errors.New("ovf file couldn't be found")
	}

	return file_matches[0], err
}

func (p *VBoxProvider) renameOVF(dir string) error {
	log.Println("Looking for OVF to rename...")
	ovf, err := p.findOvf(dir)
	if err != nil {
		return err
	}

	log.Printf("Renaming: '%s' => box.ovf", ovf)
	return os.Rename(ovf, filepath.Join(dir, "box.ovf"))
}

func (p *VBoxProvider) findBaseMacAddress(dir string) (string, error) {
	log.Println("Looking for OVF for base mac address...")
	ovf, err := p.findOvf(dir)
	if err != nil {
		return "", err
	}

	f, err := os.Open(ovf)
	if err != nil {
		return "", err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`<Adapter slot="0".+?MACAddress="(.+?)"`)
	matches := re.FindSubmatch(data)
	if matches == nil {
		return "", errors.New("can't find base mac address in OVF")
	}

	log.Printf("Base mac address: %s", string(matches[1]))
	return string(matches[1]), nil
}

// DecompressOva takes an ova file and decompresses it into the target
// directory.
func DecompressOva(dir, src string) error {
	log.Printf("Turning ova to dir: %s => %s", src, dir)
	srcF, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcF.Close()

	tarReader := tar.NewReader(srcF)
	for {
		hdr, err := tarReader.Next()
		if hdr == nil || err == io.EOF {
			break
		}

		info := hdr.FileInfo()

		// Shouldn't be any directories, skip them
		if info.IsDir() {
			continue
		}

		// We wrap this in an anonymous function so that the defers
		// inside are handled more quickly so we can give up file handles.
		err = func() error {
			path := filepath.Join(dir, info.Name())
			output, err := os.Create(path)
			if err != nil {
				return err
			}
			defer output.Close()

			os.Chmod(path, info.Mode())
			os.Chtimes(path, hdr.AccessTime, hdr.ModTime)
			_, err = io.Copy(output, tarReader)
			return err
		}()
		if err != nil {
			return err
		}
	}

	return nil
}

var vboxVagrantfile = `
Vagrant.configure("2") do |config|
  config.vm.base_mac = "%s"
end
`
