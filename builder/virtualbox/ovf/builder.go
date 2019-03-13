package ovf

import (
	"errors"
	"fmt"
	"log"

	vboxcommon "github.com/hashicorp/packer/builder/virtualbox/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
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
func (b *Builder) Run(ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	// Create the driver that we'll use to communicate with VirtualBox
	driver, err := vboxcommon.NewDriver()
	if err != nil {
		return nil, fmt.Errorf("Failed creating VirtualBox driver: %s", err)
	}

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("debug", b.config.PackerDebug)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps.
	steps := []multistep.Step{
		&common.StepOutputDir{
			Force: b.config.PackerForce,
			Path:  b.config.OutputDir,
		},
		new(vboxcommon.StepSuppressMessages),
		&common.StepCreateFloppy{
			Files:       b.config.FloppyConfig.FloppyFiles,
			Directories: b.config.FloppyConfig.FloppyDirectories,
		},
		&common.StepHTTPServer{
			HTTPDir:     b.config.HTTPDir,
			HTTPPortMin: b.config.HTTPPortMin,
			HTTPPortMax: b.config.HTTPPortMax,
		},
		&vboxcommon.StepSshKeyPair{
			Debug:        b.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("%s.pem", b.config.PackerBuildName),
			Comm:         &b.config.Comm,
		},
		&vboxcommon.StepDownloadGuestAdditions{
			GuestAdditionsMode:   b.config.GuestAdditionsMode,
			GuestAdditionsURL:    b.config.GuestAdditionsURL,
			GuestAdditionsSHA256: b.config.GuestAdditionsSHA256,
			Ctx:                  b.config.ctx,
		},
		&common.StepDownload{
			Checksum:     b.config.Checksum,
			ChecksumType: b.config.ChecksumType,
			Description:  "OVF/OVA",
			Extension:    "ova",
			ResultKey:    "vm_path",
			TargetPath:   b.config.TargetPath,
			Url:          []string{b.config.SourcePath},
		},
		&StepImport{
			Name:        b.config.VMName,
			ImportFlags: b.config.ImportFlags,
		},
		&vboxcommon.StepAttachGuestAdditions{
			GuestAdditionsMode:      b.config.GuestAdditionsMode,
			GuestAdditionsInterface: b.config.GuestAdditionsInterface,
		},
		&vboxcommon.StepConfigureVRDP{
			VRDPBindAddress: b.config.VRDPBindAddress,
			VRDPPortMin:     b.config.VRDPPortMin,
			VRDPPortMax:     b.config.VRDPPortMax,
		},
		new(vboxcommon.StepAttachFloppy),
		&vboxcommon.StepForwardSSH{
			CommConfig:     &b.config.SSHConfig.Comm,
			HostPortMin:    b.config.SSHHostPortMin,
			HostPortMax:    b.config.SSHHostPortMax,
			SkipNatMapping: b.config.SSHSkipNatMapping,
		},
		&vboxcommon.StepVBoxManage{
			Commands: b.config.VBoxManage,
			Ctx:      b.config.ctx,
		},
		&vboxcommon.StepRun{
			Headless: b.config.Headless,
		},
		&vboxcommon.StepTypeBootCommand{
			BootWait:      b.config.BootWait,
			BootCommand:   b.config.FlatBootCommand(),
			VMName:        b.config.VMName,
			Ctx:           b.config.ctx,
			GroupInterval: b.config.BootConfig.BootGroupInterval,
			Comm:          &b.config.Comm,
		},
		&communicator.StepConnect{
			Config:    &b.config.SSHConfig.Comm,
			Host:      vboxcommon.CommHost(b.config.SSHConfig.Comm.SSHHost),
			SSHConfig: b.config.SSHConfig.Comm.SSHConfigFunc(),
			SSHPort:   vboxcommon.SSHPort,
			WinRMPort: vboxcommon.SSHPort,
		},
		&vboxcommon.StepUploadVersion{
			Path: *b.config.VBoxVersionFile,
		},
		&vboxcommon.StepUploadGuestAdditions{
			GuestAdditionsMode: b.config.GuestAdditionsMode,
			GuestAdditionsPath: b.config.GuestAdditionsPath,
			Ctx:                b.config.ctx,
		},
		new(common.StepProvision),
		&common.StepCleanupTempKeys{
			Comm: &b.config.SSHConfig.Comm,
		},
		&vboxcommon.StepShutdown{
			Command: b.config.ShutdownCommand,
			Timeout: b.config.ShutdownTimeout,
			Delay:   b.config.PostShutdownDelay,
		},
		&vboxcommon.StepRemoveDevices{
			GuestAdditionsInterface: b.config.GuestAdditionsInterface,
		},
		&vboxcommon.StepVBoxManage{
			Commands: b.config.VBoxManagePost,
			Ctx:      b.config.ctx,
		},
		&vboxcommon.StepExport{
			Format:         b.config.Format,
			OutputDir:      b.config.OutputDir,
			ExportOpts:     b.config.ExportOpts.ExportOpts,
			SkipNatMapping: b.config.SSHSkipNatMapping,
			SkipExport:     b.config.SkipExport,
		},
	}

	// Run the steps.
	b.runner = common.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
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
