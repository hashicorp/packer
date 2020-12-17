package vagrant

import (
	"bytes"
	"text/template"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type googleVagrantfileTemplate struct {
	Image string ""
}

type GoogleProvider struct{}

func (p *GoogleProvider) KeepInputArtifact() bool {
	return true
}

func (p *GoogleProvider) Process(ui packersdk.Ui, artifact packersdk.Artifact, dir string) (vagrantfile string, metadata map[string]interface{}, err error) {
	// Create the metadata
	metadata = map[string]interface{}{"provider": "google"}

	// Build up the template data to build our Vagrantfile
	tplData := &googleVagrantfileTemplate{}
	tplData.Image = artifact.Id()

	// Build up the Vagrantfile
	var contents bytes.Buffer
	t := template.Must(template.New("vf").Parse(defaultGoogleVagrantfile))
	err = t.Execute(&contents, tplData)
	vagrantfile = contents.String()
	return
}

var defaultGoogleVagrantfile = `
Vagrant.configure("2") do |config|
  config.vm.provider :google do |google|
    google.image = "{{ .Image }}"
  end
end
`
