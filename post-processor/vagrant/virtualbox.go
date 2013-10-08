package vagrant

import (
	"archive/tar"
	"compress/flate"
	"errors"
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

type VBoxBoxConfig struct {
	common.PackerConfig `mapstructure:",squash"`

	OutputPath          string `mapstructure:"output"`
	VagrantfileTemplate string `mapstructure:"vagrantfile_template"`
	CompressionLevel    string `mapstructure:"compression_level"`

	tpl *packer.ConfigTemplate
}

type VBoxVagrantfileTemplate struct {
	BaseMacAddress string
}

type VBoxBoxPostProcessor struct {
	config VBoxBoxConfig
}

func (p *VBoxBoxPostProcessor) Configure(raws ...interface{}) error {
	md, err := common.DecodeConfig(&p.config, raws...)
	if err != nil {
		return err
	}

	p.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return err
	}
	p.config.tpl.UserVars = p.config.PackerUserVars

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)

	validates := map[string]*string{
		"output":               &p.config.OutputPath,
		"vagrantfile_template": &p.config.VagrantfileTemplate,
		"compression_level":    &p.config.CompressionLevel,
	}

	for n, ptr := range validates {
		if err := p.config.tpl.Validate(*ptr); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error parsing %s: %s", n, err))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *VBoxBoxPostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	var err error

	// Compile the output path
	outputPath, err := p.config.tpl.Process(p.config.OutputPath, &OutputPathTemplate{
		ArtifactId: artifact.Id(),
		BuildName:  p.config.PackerBuildName,
		Provider:   "virtualbox",
	})
	if err != nil {
		return nil, false, err
	}

	// Create a temporary directory for us to build the contents of the box in
	dir, err := ioutil.TempDir("", "packer")
	if err != nil {
		return nil, false, err
	}
	defer os.RemoveAll(dir)

	// Copy all of the original contents into the temporary directory
	for _, path := range artifact.Files() {

		// We treat OVA files specially, we unpack those into the temporary
		// directory so we can get the resulting disk and OVF.
		if extension := filepath.Ext(path); extension == ".ova" {
			ui.Message(fmt.Sprintf("Unpacking OVA: %s", path))
			if err := DecompressOva(dir, filepath.Base(path)); err != nil {
				return nil, false, err
			}
		} else {
			ui.Message(fmt.Sprintf("Copying: %s", path))
			dstPath := filepath.Join(dir, filepath.Base(path))
			if err := CopyContents(dstPath, path); err != nil {
				return nil, false, err
			}
		}

	}

	// Create the Vagrantfile from the template
	tplData := &VBoxVagrantfileTemplate{}
	tplData.BaseMacAddress, err = p.findBaseMacAddress(dir)
	if err != nil {
		return nil, false, err
	}

	vf, err := os.Create(filepath.Join(dir, "Vagrantfile"))
	if err != nil {
		return nil, false, err
	}
	defer vf.Close()

	vagrantfileContents := defaultVBoxVagrantfile
	if p.config.VagrantfileTemplate != "" {
		f, err := os.Open(p.config.VagrantfileTemplate)
		if err != nil {
			return nil, false, err
		}
		defer f.Close()

		contents, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, false, err
		}

		vagrantfileContents = string(contents)
	}

	vagrantfileContents, err = p.config.tpl.Process(vagrantfileContents, tplData)
	if err != nil {
		return nil, false, fmt.Errorf("Error writing Vagrantfile: %s", err)
	}
	vf.Write([]byte(vagrantfileContents))
	vf.Close()

	var level int = flate.DefaultCompression
	if p.config.CompressionLevel != "" {
		level, err = strconv.Atoi(p.config.CompressionLevel)
		if err != nil {
			return nil, false, err
		}
	}

	// Create the metadata
	metadata := map[string]string{"provider": "virtualbox"}
	if err := WriteMetadata(dir, metadata); err != nil {
		return nil, false, err
	}

	// Rename the OVF file to box.ovf, as required by Vagrant
	ui.Message("Renaming the OVF to box.ovf...")
	if err := p.renameOVF(dir); err != nil {
		return nil, false, err
	}

	// Compress the directory to the given output path
	ui.Message(fmt.Sprintf("Compressing box..."))
	if err := DirToBox(outputPath, dir, ui, level); err != nil {
		return nil, false, err
	}

	return NewArtifact("virtualbox", outputPath), false, nil
}

func (p *VBoxBoxPostProcessor) findOvf(dir string) (string, error) {
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

func (p *VBoxBoxPostProcessor) renameOVF(dir string) error {
	log.Println("Looking for OVF to rename...")
	ovf, err := p.findOvf(dir)
	if err != nil {
		return err
	}

	log.Printf("Renaming: '%s' => box.ovf", ovf)
	return os.Rename(ovf, filepath.Join(dir, "box.ovf"))
}

func (p *VBoxBoxPostProcessor) findBaseMacAddress(dir string) (string, error) {
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

var defaultVBoxVagrantfile = `
Vagrant.configure("2") do |config|
config.vm.base_mac = "{{ .BaseMacAddress }}"
end
`
