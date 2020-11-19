package clone

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/builder/vsphere/common"
	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	warnings, errs := b.config.Prepare(raws...)
	if errs != nil {
		return nil, warnings, errs
	}

	return nil, warnings, nil
}

func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packer.Hook) (packersdk.Artifact, error) {
	state := new(multistep.BasicStateBag)
	state.Put("debug", b.config.PackerDebug)
	state.Put("hook", hook)
	state.Put("ui", ui)

	var steps []multistep.Step

	steps = append(steps,
		&common.StepConnect{
			Config: &b.config.ConnectConfig,
		},
		&commonsteps.StepCreateCD{
			Files: b.config.CDConfig.CDFiles,
			Label: b.config.CDConfig.CDLabel,
		},
		&common.StepRemoteUpload{
			Datastore:                  b.config.Datastore,
			Host:                       b.config.Host,
			SetHostForDatastoreUploads: b.config.SetHostForDatastoreUploads,
		},
		&StepCloneVM{
			Config:   &b.config.CloneConfig,
			Location: &b.config.LocationConfig,
			Force:    b.config.PackerConfig.PackerForce,
		},
		&common.StepConfigureHardware{
			Config: &b.config.HardwareConfig,
		},
		&common.StepAddCDRom{
			Config: &b.config.CDRomConfig,
		},
		&common.StepConfigParams{
			Config: &b.config.ConfigParamsConfig,
		},
	)

	if b.config.CustomizeConfig != nil {
		steps = append(steps, &StepCustomize{
			Config: b.config.CustomizeConfig,
		})
	}

	if b.config.Comm.Type != "none" {
		steps = append(steps,
			&commonsteps.StepCreateFloppy{
				Files:       b.config.FloppyFiles,
				Directories: b.config.FloppyDirectories,
				Label:       b.config.FloppyLabel,
			},
			&common.StepAddFloppy{
				Config:                     &b.config.FloppyConfig,
				Datastore:                  b.config.Datastore,
				Host:                       b.config.Host,
				SetHostForDatastoreUploads: b.config.SetHostForDatastoreUploads,
			},
			&common.StepHTTPIPDiscover{
				HTTPIP:  b.config.BootConfig.HTTPIP,
				Network: b.config.WaitIpConfig.GetIPNet(),
			},
			&commonsteps.StepHTTPServer{
				HTTPDir:     b.config.HTTPDir,
				HTTPPortMin: b.config.HTTPPortMin,
				HTTPPortMax: b.config.HTTPPortMax,
				HTTPAddress: b.config.HTTPAddress,
			},
			&common.StepSshKeyPair{
				Debug:        b.config.PackerDebug,
				DebugKeyPath: fmt.Sprintf("%s.pem", b.config.PackerBuildName),
				Comm:         &b.config.Comm,
			},
			&common.StepRun{
				Config:   &b.config.RunConfig,
				SetOrder: false,
			},
			&common.StepBootCommand{
				Config: &b.config.BootConfig,
				Ctx:    b.config.ctx,
				VMName: b.config.VMName,
			},
			&common.StepWaitForIp{
				Config: &b.config.WaitIpConfig,
			},
			&communicator.StepConnect{
				Config:    &b.config.Comm,
				Host:      common.CommHost(b.config.Comm.Host()),
				SSHConfig: b.config.Comm.SSHConfigFunc(),
			},
			&commonsteps.StepProvision{},
			&common.StepShutdown{
				Config: &b.config.ShutdownConfig,
			},
			&common.StepRemoveFloppy{
				Datastore: b.config.Datastore,
				Host:      b.config.Host,
			},
		)
	}

	steps = append(steps,
		&common.StepRemoveCDRom{
			Config: &b.config.RemoveCDRomConfig,
		},
		&common.StepCreateSnapshot{
			CreateSnapshot: b.config.CreateSnapshot,
		},
		&common.StepConvertToTemplate{
			ConvertToTemplate: b.config.ConvertToTemplate,
		},
	)

	if b.config.ContentLibraryDestinationConfig != nil {
		steps = append(steps, &common.StepImportToContentLibrary{
			ContentLibConfig: b.config.ContentLibraryDestinationConfig,
		})
	}

	if b.config.Export != nil {
		steps = append(steps, &common.StepExport{
			Name:      b.config.Export.Name,
			Force:     b.config.Export.Force,
			Images:    b.config.Export.Images,
			Manifest:  b.config.Export.Manifest,
			OutputDir: b.config.Export.OutputDir.OutputDir,
			Options:   b.config.Export.Options,
		})
	}

	b.runner = commonsteps.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
	b.runner.Run(ctx, state)

	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	if _, ok := state.GetOk("vm"); !ok {
		return nil, nil
	}
	artifact := &common.Artifact{
		Name:      b.config.VMName,
		VM:        state.Get("vm").(*driver.VirtualMachineDriver),
		StateData: map[string]interface{}{"generated_data": state.Get("generated_data")},
	}
	if b.config.Export != nil {
		artifact.Outconfig = &b.config.Export.OutputDir
	}

	return artifact, nil
}
