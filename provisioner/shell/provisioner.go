// This package implements a provisioner for Packer that executes
// shell scripts within the remote machine.
package shell

import (
	"github.com/mitchellh/packer/packer"
)

// TODO(mitchellh): config
type config struct {
}

type Provisioner struct {
	config config
}

func (p *Provisioner) Prepare(raw interface{}, ui packer.Ui) {
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) {
	ui.Say("PROVISIONING SOME STUFF")
}
