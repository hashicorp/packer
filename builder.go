package main

import (
	"errors"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"github.com/mitchellh/multistep"
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
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)

	var steps []multistep.Step

	steps = append(steps,
		&StepConnect{
			config: &b.config.ConnectConfig,
		},
		&StepCloneVM{
			config: &b.config.CloneConfig,
		},
		&StepConfigureHardware{
			config: &b.config.HardwareConfig,
		},
	)

	if b.config.Comm.Type != "none" {
		steps = append(steps,
			&StepRun{},
			&communicator.StepConnect{
				Config:    &b.config.Comm,
				Host:      commHost,
				SSHConfig: sshConfig,
			},
			&common.StepProvision{},
			&StepShutdown{
				config: &b.config.ShutdownConfig,
			},
		)
	}

	steps = append(steps,
		&StepCreateSnapshot{
			createSnapshot: b.config.CreateSnapshot,
		},
		&StepConvertToTemplate{
			ConvertToTemplate: b.config.ConvertToTemplate,
		},
	)

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

	artifact := &Artifact{
		Name: b.config.VMName,
		VM:   state.Get("vm").(*driver.VirtualMachine),
	}
	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		b.runner.Cancel()
	}
}
