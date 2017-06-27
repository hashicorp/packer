package main

import (
	"errors"
	"log"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"github.com/hashicorp/packer/helper/communicator"
	gossh "golang.org/x/crypto/ssh"
	"github.com/hashicorp/packer/communicator/ssh"
)

type Builder struct {
	config *Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	c, warnings, errs := NewConfig(raws...)
	if errs != nil {
		return warnings, errs
	}
	b.config = c

	return warnings, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps.
	steps := []multistep.Step{
		&StepSetup{
			config: b.config,
		},
		&StepCloneVM{
			config: b.config,
		},
		&StepConfigureHW{
			config: b.config,
		},
		&StepRun{},
		&communicator.StepConnect{
			Config:    &b.config.Config,
			Host:      func(state multistep.StateBag) (string, error) {
				return state.Get("ip").(string), nil
			},
			SSHConfig: func(multistep.StateBag) (*gossh.ClientConfig, error) {
				return &gossh.ClientConfig{
					User: b.config.Config.SSHUsername,
					Auth: []gossh.AuthMethod{
						gossh.Password(b.config.Config.SSHPassword),
						gossh.KeyboardInteractive(
							ssh.PasswordKeyboardInteractive(b.config.Config.SSHPassword)),
					},
					// TODO: add a proper verification
					HostKeyCallback: gossh.InsecureIgnoreHostKey(),
				}, nil
			},
		},
		&common.StepProvision{},
		&StepShutdown{
			Command:    b.config.ShutdownCommand,
			ShutdownTimeout: b.config.ShutdownTimeout,
		},
		&StepCreateSnapshot{
			createSnapshot: b.config.CreateSnapshot,
		},
		&StepConvertToTemplate{
			ConvertToTemplate: b.config.ConvertToTemplate,
		},
	}

	// Run!
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If we were interrupted or cancelled, then just exit.
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, errors.New("Build was cancelled.")
	}

	if _, ok := state.GetOk(multistep.StateHalted); ok {
		return nil, errors.New("Build was halted.")
	}

	// No errors, must've worked
	artifact := &Artifact{
		VMName: b.config.VMName,
	}
	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
