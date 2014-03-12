package ovf

import (
	"errors"
	"fmt"
	"log"

	"github.com/mitchellh/multistep"
	vboxcommon "github.com/mitchellh/packer/builder/virtualbox/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
)

// Builder implements packer.Builder and builds the actual VirtualBox
// images.
type Builder struct {
	config *Config
	runner multistep.Runner
}

// Prepare processes the build configuration parameters.
func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	c, warnings, errs := NewConfig(raws...)
	if errs != nil {
		return warnings, errs
	}
	b.config = c

	return warnings, nil
}

// Run executes a Packer build and returns a packer.Artifact representing
// a VirtualBox appliance.
func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	// Create the driver that we'll use to communicate with VirtualBox
	driver, err := vboxcommon.NewDriver()
	if err != nil {
		return nil, fmt.Errorf("Failed creating VirtualBox driver: %s", err)
	}

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps.
	steps := []multistep.Step{
		&vboxcommon.StepOutputDir{
			Force: b.config.PackerForce,
			Path:  b.config.OutputDir,
		},
		new(vboxcommon.StepSuppressMessages),
		&common.StepCreateFloppy{
			Files: b.config.FloppyFiles,
		},
		&StepImport{
			Name:       b.config.VMName,
			SourcePath: b.config.SourcePath,
			ImportOpts: b.config.ImportOpts,
		},
		/*
			new(stepAttachGuestAdditions),
		*/
		new(vboxcommon.StepAttachFloppy),
		&vboxcommon.StepForwardSSH{
			GuestPort:   b.config.SSHPort,
			HostPortMin: b.config.SSHHostPortMin,
			HostPortMax: b.config.SSHHostPortMax,
		},
		&vboxcommon.StepVBoxManage{
			Commands: b.config.VBoxManage,
			Tpl:      b.config.tpl,
		},
		&vboxcommon.StepRun{
			BootWait: b.config.BootWait,
			Headless: b.config.Headless,
		},
		&common.StepConnectSSH{
			SSHAddress:     vboxcommon.SSHAddress,
			SSHConfig:      vboxcommon.SSHConfigFunc(b.config.SSHConfig),
			SSHWaitTimeout: b.config.SSHWaitTimeout,
		},
		&vboxcommon.StepUploadVersion{
			Path: b.config.VBoxVersionFile,
		},
		/*
			new(stepUploadGuestAdditions),
		*/
		new(common.StepProvision),
		&vboxcommon.StepShutdown{
			Command: b.config.ShutdownCommand,
			Timeout: b.config.ShutdownTimeout,
		},
		new(vboxcommon.StepRemoveDevices),
		&vboxcommon.StepExport{
			Format:     b.config.Format,
			OutputDir:  b.config.OutputDir,
			ExportOpts: b.config.ExportOpts.ExportOpts,
		},
	}

	// Run the steps.
	if b.config.PackerDebug {
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.runner = &multistep.BasicRunner{Steps: steps}
	}
	b.runner.Run(state)

	// Report any errors.
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

	return vboxcommon.NewArtifact(b.config.OutputDir)
}

// Cancel.
func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
