package vagrant

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type AWSProvider struct{}

func (p *AWSProvider) KeepInputArtifact() bool {
	return true
}

func (p *AWSProvider) Process(ui packersdk.Ui, artifact packer.Artifact, dir string) (vagrantfile string, metadata map[string]interface{}, err error) {
	// Create the metadata
	metadata = map[string]interface{}{"provider": "aws"}

	// Build up the template data to build our Vagrantfile
	tplData := &awsVagrantfileTemplate{
		Images: make(map[string]string),
	}

	for _, regions := range strings.Split(artifact.Id(), ",") {
		parts := strings.Split(regions, ":")
		if len(parts) != 2 {
			err = fmt.Errorf("Poorly formatted artifact ID: %s", artifact.Id())
			return
		}

		tplData.Images[parts[0]] = parts[1]
	}

	// Build up the contents
	var contents bytes.Buffer
	t := template.Must(template.New("vf").Parse(defaultAWSVagrantfile))
	err = t.Execute(&contents, tplData)
	vagrantfile = contents.String()
	return
}

type awsVagrantfileTemplate struct {
	Images map[string]string
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
