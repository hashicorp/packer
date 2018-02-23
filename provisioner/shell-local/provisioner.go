package shell

import (
	"errors"
	"fmt"

	"github.com/hashicorp/packer/common"
	sl "github.com/hashicorp/packer/common/shell-local"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Command is the command to execute
	Command string

	// ExecuteCommand is the command used to execute the command.
	ExecuteCommand []string `mapstructure:"execute_command"`

	ctx interpolate.Context
}

type Provisioner struct {
	config Config
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"execute_command",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	var errs *packer.MultiError
	if p.config.Command == "" {
		errs = packer.MultiErrorAppend(errs,
			errors.New("command must be specified"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, _ packer.Communicator) error {
	// Make another communicator for local
	comm := &sl.Communicator{
		Ctx:            p.config.ctx,
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
