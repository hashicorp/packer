package vagrant

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type scalewayVagrantfileTemplate struct {
	Image  string ""
	Region string ""
}

type ScalewayProvider struct{}

func (p *ScalewayProvider) KeepInputArtifact() bool {
	return true
}

func (p *ScalewayProvider) Process(ui packersdk.Ui, artifact packersdk.Artifact, dir string) (vagrantfile string, metadata map[string]interface{}, err error) {
	// Create the metadata
	metadata = map[string]interface{}{"provider": "scaleway"}

	// Determine the image and region...
	tplData := &scalewayVagrantfileTemplate{}

	parts := strings.Split(artifact.Id(), ":")
	if len(parts) != 2 {
		err = fmt.Errorf("Poorly formatted artifact ID: %s", artifact.Id())
		return
	}
	tplData.Region = parts[0]
	tplData.Image = parts[1]

	// Build up the Vagrantfile
	var contents bytes.Buffer
	t := template.Must(template.New("vf").Parse(defaultScalewayVagrantfile))
	err = t.Execute(&contents, tplData)
	vagrantfile = contents.String()
	return
}

var defaultScalewayVagrantfile = `
Vagrant.configure("2") do |config|
  config.vm.provider :scaleway do |scaleway|
	scaleway.image = "{{ .Image }}"
	scaleway.region = "{{ .Region }}"
  end
end
`
