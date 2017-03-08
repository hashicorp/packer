package iso

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/mitchellh/multistep"
	vspcommon "github.com/mitchellh/packer/builder/vsphere/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/packer"
)

// Builder implements packer.Builder and builds the actual VMware
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

// Run Launch the build steps
func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	driver, err := vspcommon.NewDriver(&b.config.DriverConfig, &b.config.SSHConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed creating VSphere driver: %s", err)
	}

	// Setup the directory
	dir := new(vspcommon.LocalOutputDir)
	dir.SetOutputDir(b.config.OutputDir)

	// Setup the state bag
	state := new(multistep.BasicStateBag)
	state.Put("cache", cache)
	state.Put("config", &b.config)
	state.Put("debug", b.config.PackerDebug)
	state.Put("dir", dir)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps.
	steps := []multistep.Step{
		&common.StepDownload{
			Checksum:     b.config.ISOChecksum,
			ChecksumType: b.config.ISOChecksumType,
			Description:  "ISO",
			Extension:    b.config.TargetExtension,
			ResultKey:    "iso_path",
			TargetPath:   b.config.TargetPath,
			Url:          b.config.ISOUrls,
		},
		&vspcommon.StepOutputDir{
			Force: b.config.PackerForce,
		},
		&common.StepCreateFloppy{
			Files:       b.config.FloppyConfig.FloppyFiles,
			Directories: b.config.FloppyConfig.FloppyDirectories,
		},
		&vspcommon.StepRemoteUpload{
			Key:     "floppy_path",
			Message: "Uploading Floppy to remote machine...",
		},
		&vspcommon.StepRemoteUpload{
			Key:     "iso_path",
			Message: "Uploading ISO to remote machine...",
		},
		&stepCreateVM{
			VMName:             b.config.VMName,
			Folder:             b.config.RemoteFolder,
			Datastore:          b.config.RemoteDatastore,
			Cpu:                b.config.Cpu,
			MemSize:            b.config.MemSize,
			DiskSize:           b.config.DiskSize,
			AdditionalDiskSize: b.config.AdditionalDiskSize,
			DiskThick:          b.config.DiskThick,
			GuestType:          b.config.GuestOSType,
			NetworkName:        b.config.NetworkName,
			NetworkAdapter:     b.config.NetworkAdapter,
			Annotation:         b.config.Annotation,
		},
		&vspcommon.StepRegister{
			Format:         b.config.Format,
			KeepRegistered: b.config.KeepRegistered,
		},
		&vspcommon.StepConfigureVM{
			CustomData: b.config.VMXData,
		},
		&common.StepHTTPServer{
			HTTPDir:     b.config.HTTPDir,
			HTTPPortMin: b.config.HTTPPortMin,
			HTTPPortMax: b.config.HTTPPortMax,
		},
		&vspcommon.StepConfigureVNC{
			VNCBindAddress:     b.config.VNCBindAddress,
			VNCPortMin:         b.config.VNCPortMin,
			VNCPortMax:         b.config.VNCPortMax,
			VNCDisablePassword: b.config.VNCDisablePassword,
		},
		&vspcommon.StepRun{
			BootWait:           b.config.BootWait,
			DurationBeforeStop: 5 * time.Second,
		},
		&vspcommon.StepTypeBootCommand{
			BootCommand: b.config.BootCommand,
			VMName:      b.config.VMName,
			Ctx:         b.config.ctx,
		},
		&communicator.StepConnect{
			Config:    &b.config.SSHConfig.Comm,
			Host:      vspcommon.CommHost,
			SSHConfig: vspcommon.SSHConfigFunc(&b.config.SSHConfig),
		},
		&vspcommon.StepUploadTools{},
		&common.StepProvision{},
		&vspcommon.StepShutdown{
			Command: b.config.ShutdownCommand,
			Timeout: b.config.ShutdownTimeout,
		},
		&vspcommon.StepCleanVM{
			CustomData: b.config.VMXDataPost,
		},
		&vspcommon.StepExport{
			Format:         b.config.Format,
			OutputPath:     b.config.OutputDir,
			SkipExport:     b.config.SkipExport,
			OVFToolOptions: b.config.OVFToolOptions,
		},
	}

	// Run!
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

	//TODO: When SkipExport=true this list is empty (No post-processor can be runned ?)
	// Compile the artifact list
	var files []string
	if b.config.Format != "" {
		dir = new(vspcommon.LocalOutputDir)
		dir.SetOutputDir(b.config.OutputDir)
		files, err = dir.ListFiles()
	} else {
		files, err = state.Get("dir").(vspcommon.OutputDir).ListFiles()
	}
	if err != nil {
		return nil, err
	}

	return vspcommon.NewArtifact(dir, files), nil
}

//Cancel
func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
