package vagrant

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/packer"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

type VBoxBoxConfig struct {
	OutputPath          string `mapstructure:"output"`
	VagrantfileTemplate string `mapstructure:"vagrantfile_template"`
}

type VBoxVagrantfileTemplate struct {
	BaseMacAddress string
}

type VBoxBoxPostProcessor struct {
	config VBoxBoxConfig
}

func (p *VBoxBoxPostProcessor) Configure(raw interface{}) error {
	err := mapstructure.Decode(raw, &p.config)
	if err != nil {
		return err
	}

	return nil
}

func (p *VBoxBoxPostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, error) {
	// TODO(mitchellh): Actually parse the base mac address
	tplData := &VBoxVagrantfileTemplate{}

	// Compile the output path
	outputPath, err := ProcessOutputPath(p.config.OutputPath, "virtualbox", artifact)
	if err != nil {
		return nil, err
	}

	// Create a temporary directory for us to build the contents of the box in
	dir, err := ioutil.TempDir("", "packer")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	// Copy all of the original contents into the temporary directory
	for _, path := range artifact.Files() {
		ui.Message(fmt.Sprintf("Copying: %s", path))
		src, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer src.Close()

		dst, err := os.Create(filepath.Join(dir, filepath.Base(path)))
		if err != nil {
			return nil, err
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return nil, err
		}
	}

	// Create the Vagrantfile from the template
	vf, err := os.Create(filepath.Join(dir, "Vagrantfile"))
	if err != nil {
		return nil, err
	}
	defer vf.Close()

	vagrantfileContents := defaultVBoxVagrantfile
	if p.config.VagrantfileTemplate != "" {
		f, err := os.Open(p.config.VagrantfileTemplate)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		contents, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}

		vagrantfileContents = string(contents)
	}

	t := template.Must(template.New("vagrantfile").Parse(vagrantfileContents))
	t.Execute(vf, tplData)
	vf.Close()

	// Create the metadata
	metadata := map[string]string{"provider": "virtualbox"}
	if err := WriteMetadata(dir, metadata); err != nil {
		return nil, err
	}

	// Compress the directory to the given output path
	ui.Message(fmt.Sprintf("Compressing box..."))
	if err := DirToBox(outputPath, dir); err != nil {
		return nil, err
	}

	return NewArtifact("virtualbox", outputPath), nil
}

var defaultVBoxVagrantfile = `
Vagrant.configure("2") do |config|
  config.vm.provider "virtualbox" do |vb|
    # TODO
  end
end
`
