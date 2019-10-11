package vagrant

import (
	"fmt"

	"github.com/hashicorp/packer/packer"
)

type DockerProvider struct{}

func (p *DockerProvider) KeepInputArtifact() bool {
	return false
}

func (p *DockerProvider) Process(ui packer.Ui, artifact packer.Artifact, dir string) (vagrantfile string, metadata map[string]interface{}, err error) {
	// Create the metadata
	metadata = map[string]interface{}{"provider": "docker"}

	vagrantfile = fmt.Sprintf(dockerVagrantfile, artifact.Id())
	return
}

var dockerVagrantfile = `
Vagrant.configure("2") do |config|
	config.vm.provider :docker do |docker, override|
		docker.image = "%s"
	end
end
`
