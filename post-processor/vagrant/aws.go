package vagrant

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type AWSBoxConfig struct {
	OutputPath          string `mapstructure:"output"`
	VagrantfileTemplate string `mapstructure:"vagrantfile_template"`

	PackerBuildName string `mapstructure:"packer_build_name"`
}

type AWSVagrantfileTemplate struct {
	Images map[string]string
}

type AWSBoxPostProcessor struct {
	config AWSBoxConfig
}

func (p *AWSBoxPostProcessor) Configure(raws ...interface{}) error {
	for _, raw := range raws {
		err := mapstructure.Decode(raw, &p.config)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *AWSBoxPostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	// Determine the regions...
	tplData := &AWSVagrantfileTemplate{
		Images: make(map[string]string),
	}

	for _, regions := range strings.Split(artifact.Id(), ",") {
		parts := strings.Split(regions, ":")
		if len(parts) != 2 {
			return nil, false, fmt.Errorf("Poorly formatted artifact ID: %s", artifact.Id())
		}

		tplData.Images[parts[0]] = parts[1]
	}

	// Compile the output path
	outputPath, err := ProcessOutputPath(p.config.OutputPath,
		p.config.PackerBuildName, "aws", artifact)
	if err != nil {
		return nil, false, err
	}

	// Create a temporary directory for us to build the contents of the box in
	dir, err := ioutil.TempDir("", "packer")
	if err != nil {
		return nil, false, err
	}
	defer os.RemoveAll(dir)

	// Create the Vagrantfile from the template
	vf, err := os.Create(filepath.Join(dir, "Vagrantfile"))
	if err != nil {
		return nil, false, err
	}
	defer vf.Close()

	vagrantfileContents := defaultAWSVagrantfile
	if p.config.VagrantfileTemplate != "" {
		log.Printf("Using vagrantfile template: %s", p.config.VagrantfileTemplate)
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

	t := template.Must(template.New("vagrantfile").Parse(vagrantfileContents))
	t.Execute(vf, tplData)
	vf.Close()

	// Create the metadata
	metadata := map[string]string{"provider": "aws"}
	if err := WriteMetadata(dir, metadata); err != nil {
		return nil, false, err
	}

	// Compress the directory to the given output path
	if err := DirToBox(outputPath, dir); err != nil {
		return nil, false, err
	}

	return NewArtifact("aws", outputPath), true, nil
}

var defaultAWSVagrantfile = `
Vagrant.configure("2") do |config|
  config.vm.provider "aws" do |aws|
    {{ range $region, $ami := .Images }}
	aws.region_config "{{ $region }}", ami: "{{ $ami }}"
	{{ end }}
  end
end
`
