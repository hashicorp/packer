package shell

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
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

	if len(p.config.ExecuteCommand) == 0 {
		if runtime.GOOS == "windows" {
			p.config.ExecuteCommand = []string{
				"cmd",
				"/C",
				"{{.Command}}",
			}
		} else {
			p.config.ExecuteCommand = []string{
				"/bin/sh",
				"-c",
				"{{.Command}}",
			}
		}
	}

	var errs *packer.MultiError
	if p.config.Command == "" {
		errs = packer.MultiErrorAppend(errs,
			errors.New("command must be specified"))
	}

	if len(p.config.ExecuteCommand) == 0 {
		errs = packer.MultiErrorAppend(errs,
			errors.New("execute_command must not be empty"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, _ packer.Communicator) error {
	// Make another communicator for local
	comm := &Communicator{
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
