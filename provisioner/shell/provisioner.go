// This package implements a provisioner for Packer that executes
// shell scripts within the remote machine.
package shell

import (
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/packer"
)

const DefaultRemotePath = "/tmp/script.sh"

// TODO(mitchellh): config
type config struct {
	// The local path of the shell script to upload and execute.
	Path string

	// The remote path where the local shell script will be uploaded to.
	// This should be set to a writable file that is in a pre-existing directory.
	RemotePath string
}

type Provisioner struct {
	config config
}

func (p *Provisioner) Prepare(raw interface{}, ui packer.Ui) {
	// TODO: errors
	_ = mapstructure.Decode(raw, &p.config)

	if p.config.RemotePath == "" {
		p.config.RemotePath = DefaultRemotePath
	}
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) {
	ui.Say("PROVISIONING SOME STUFF")
}
