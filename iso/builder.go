package iso

import (
	packerCommon "github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/common"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"github.com/mitchellh/multistep"
	"github.com/hashicorp/packer/helper/communicator"
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
	state.Put("comm", &b.config.Comm)
	state.Put("hook", hook)
	state.Put("ui", ui)

	var steps []multistep.Step

	steps = append(steps,
		&common.StepConnect{
			Config: &b.config.ConnectConfig,
		},
		&StepCreateVM{
			Config: &b.config.CreateConfig,
		},
		&StepAddCDRom{
			Config: &b.config.CDRomConfig,
		},
		&packerCommon.StepCreateFloppy{
			Files:       b.config.FloppyFiles,
			Directories: b.config.FloppyDirectories,
		},
		&StepAddFloppy{
			Config:    &b.config.FloppyConfig,
			Datastore: b.config.Datastore,
			Host: b.config.Host,
		},
		&StepConfigParams{
			Config:    &b.config.ConfigParamsConfig,
		},
	)

	if b.config.Comm.Type != "none" {
		steps = append(steps,
			&common.StepRun{
				Config: &b.config.RunConfig,
			},
			&StepBootCommand{
				Config: &b.config.BootConfig,
			},
			&common.StepWaitForIp{},
			&communicator.StepConnect{
				Config:    &b.config.Comm,
				Host:      common.CommHost,
				SSHConfig: common.SshConfig,
			},
			&packerCommon.StepProvision{},
			&common.StepShutdown{
				Config: &b.config.ShutdownConfig,
			},
		)
	}

	steps = append(steps,
		&common.StepCreateSnapshot{
			CreateSnapshot: b.config.CreateSnapshot,
		},
		&common.StepConvertToTemplate{
			ConvertToTemplate: b.config.ConvertToTemplate,
		},
	)

	// Run!
	b.runner = packerCommon.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(state)

	if err := common.CheckRunStatus(state); err != nil {
		return nil, err
	}

	artifact := &common.Artifact{
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
