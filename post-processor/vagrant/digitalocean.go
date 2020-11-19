package vagrant

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type digitalOceanVagrantfileTemplate struct {
	Image  string ""
	Region string ""
}

type DigitalOceanProvider struct{}

func (p *DigitalOceanProvider) KeepInputArtifact() bool {
	return true
}

func (p *DigitalOceanProvider) Process(ui packersdk.Ui, artifact packer.Artifact, dir string) (vagrantfile string, metadata map[string]interface{}, err error) {
	// Create the metadata
	metadata = map[string]interface{}{"provider": "digital_ocean"}

	// Determine the image and region...
	tplData := &digitalOceanVagrantfileTemplate{}

	parts := strings.Split(artifact.Id(), ":")
	if len(parts) != 2 {
		err = fmt.Errorf("Poorly formatted artifact ID: %s", artifact.Id())
		return
	}
	tplData.Region = parts[0]
	tplData.Image = parts[1]

	// Build up the Vagrantfile
	var contents bytes.Buffer
	t := template.Must(template.New("vf").Parse(defaultDigitalOceanVagrantfile))
	err = t.Execute(&contents, tplData)
	vagrantfile = contents.String()
	return
}

var defaultDigitalOceanVagrantfile = `
Vagrant.configure("2") do |config|
  config.vm.provider :digital_ocean do |digital_ocean|
	digital_ocean.image = "{{ .Image }}"
	digital_ocean.region = "{{ .Region }}"
  end
end
`
