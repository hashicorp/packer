package iso

import (
	"context"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/builder/vsphere/common"
	"github.com/hashicorp/packer/builder/vsphere/driver"
	packerCommon "github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type Builder struct {
	config *Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	c, warnings, errs := NewConfig(raws...)
	if errs != nil {
		return warnings, nil, errs
	}
	b.config = c

	return warnings, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	state := new(multistep.BasicStateBag)
	state.Put("comm", &b.config.Comm)
	state.Put("hook", hook)
	state.Put("ui", ui)

	var steps []multistep.Step

	steps = append(steps,
		&common.StepConnect{
			Config: &b.config.ConnectConfig,
		},
	)

	if b.config.ISOUrls != nil {
		steps = append(steps,
			&packerCommon.StepDownload{
				Checksum:     b.config.ISOChecksum,
				ChecksumType: b.config.ISOChecksumType,
				Description:  "ISO",
				Extension:    b.config.TargetExtension,
				ResultKey:    "iso_path",
				TargetPath:   b.config.TargetPath,
				Url:          b.config.ISOUrls,
			},
			&StepRemoteUpload{
				Datastore: b.config.Datastore,
				Host:      b.config.Host,
			},
		)
	}

	steps = append(steps,
		&StepCreateVM{
			Config:   &b.config.CreateConfig,
			Location: &b.config.LocationConfig,
			Force:    b.config.PackerConfig.PackerForce,
		},
		&common.StepConfigureHardware{
			Config: &b.config.HardwareConfig,
		},
		&StepAddCDRom{
			Config: &b.config.CDRomConfig,
		},
		&common.StepConfigParams{
			Config: &b.config.ConfigParamsConfig,
		},
	)

	if b.config.Comm.Type != "none" {
		steps = append(steps,
			&packerCommon.StepCreateFloppy{
				Files:       b.config.FloppyFiles,
				Directories: b.config.FloppyDirectories,
			},
			&StepAddFloppy{
				Config:    &b.config.FloppyConfig,
				Datastore: b.config.Datastore,
				Host:      b.config.Host,
			},
			&packerCommon.StepHTTPServer{
				HTTPDir:     b.config.HTTPDir,
				HTTPPortMin: b.config.HTTPPortMin,
				HTTPPortMax: b.config.HTTPPortMax,
			},
			&common.StepRun{
				Config:   &b.config.RunConfig,
				SetOrder: true,
			},
			&StepBootCommand{
				Config: &b.config.BootConfig,
				Ctx:    b.config.ctx,
				VMName: b.config.VMName,
			},
			&common.StepWaitForIp{
				Config: &b.config.WaitIpConfig,
			},
			&communicator.StepConnect{
				Config:    &b.config.Comm,
				Host:      common.CommHost(b.config.Comm.SSHHost),
				SSHConfig: b.config.Comm.SSHConfigFunc(),
			},
			&packerCommon.StepProvision{},
			&common.StepShutdown{
				Config: &b.config.ShutdownConfig,
			},
			&StepRemoveFloppy{
				Datastore: b.config.Datastore,
				Host:      b.config.Host,
			},
		)
	}

	steps = append(steps,
		&StepRemoveCDRom{},
		&common.StepCreateSnapshot{
			CreateSnapshot: b.config.CreateSnapshot,
		},
		&common.StepConvertToTemplate{
			ConvertToTemplate: b.config.ConvertToTemplate,
		},
	)

	b.runner = packerCommon.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	if _, ok := state.GetOk("vm"); !ok {
		return nil, nil
	}
	artifact := &common.Artifact{
		Name: b.config.VMName,
		VM:   state.Get("vm").(*driver.VirtualMachine),
	}
	return artifact, nil
}
