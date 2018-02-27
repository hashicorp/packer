package shell

import (
	"fmt"

	sl "github.com/hashicorp/packer/common/shell-local"
	"github.com/hashicorp/packer/packer"
)

type Provisioner struct {
	config sl.Config
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := sl.Decode(&p.config, raws)
	if err != nil {
		return err
	}

	return sl.Validate(&p.config)
}

func (p *Provisioner) Provision(ui packer.Ui, _ packer.Communicator) error {
	// Make another communicator for local
	comm := &sl.Communicator{
		Ctx:            p.config.Ctx,
		ExecuteCommand: p.config.ExecuteCommand,
	}

	// Build the remote command
	cmd := &packer.RemoteCmd{Command: p.config.Command}

	ui.Say(fmt.Sprintf(
		"Executing local command: %s",
		p.config.Command))
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return fmt.Errorf(
			"Error executing command: %s\n\n"+
				"Please see output above for more information.",
			p.config.Command)
	}
	if cmd.ExitStatus != 0 {
		return fmt.Errorf(
			"Erroneous exit code %d while executing command: %s\n\n"+
				"Please see output above for more information.",
			cmd.ExitStatus,
			p.config.Command)
	}

	return nil
}

func (p *Provisioner) Cancel() {
	// Just do nothing. When the process ends, so will our provisioner
}
