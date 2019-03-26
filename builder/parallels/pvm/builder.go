package pvm

import (
	"errors"
	"fmt"
	"log"

	parallelscommon "github.com/hashicorp/packer/builder/parallels/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// Builder implements packer.Builder and builds the actual Parallels
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
// a Parallels appliance.
func (b *Builder) Run(ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	// Create the driver that we'll use to communicate with Parallels
	driver, err := parallelscommon.NewDriver()
	if err != nil {
		return nil, fmt.Errorf("Failed creating Parallels driver: %s", err)
	}

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("debug", b.config.PackerDebug)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("http_port", 0)

	// Build the steps.
	steps := []multistep.Step{
		&parallelscommon.StepPrepareParallelsTools{
			ParallelsToolsMode:   b.config.ParallelsToolsMode,
			ParallelsToolsFlavor: b.config.ParallelsToolsFlavor,
		},
		&parallelscommon.StepOutputDir{
			Force: b.config.PackerForce,
			Path:  b.config.OutputDir,
		},
		&common.StepCreateFloppy{
			Files:       b.config.FloppyConfig.FloppyFiles,
			Directories: b.config.FloppyConfig.FloppyDirectories,
		},
		&StepImport{
			Name:       b.config.VMName,
			SourcePath: b.config.SourcePath,
		},
		&parallelscommon.StepAttachParallelsTools{
			ParallelsToolsMode: b.config.ParallelsToolsMode,
		},
		new(parallelscommon.StepAttachFloppy),
		&parallelscommon.StepPrlctl{
			Commands: b.config.Prlctl,
			Ctx:      b.config.ctx,
		},
		&parallelscommon.StepRun{},
		&parallelscommon.StepTypeBootCommand{
			BootCommand:    b.config.FlatBootCommand(),
			BootWait:       b.config.BootWait,
			HostInterfaces: []string{},
			VMName:         b.config.VMName,
			Ctx:            b.config.ctx,
			GroupInterval:  b.config.BootConfig.BootGroupInterval,
		},
		&communicator.StepConnect{
			Config:    &b.config.SSHConfig.Comm,
			Host:      parallelscommon.CommHost,
			SSHConfig: b.config.SSHConfig.Comm.SSHConfigFunc(),
		},
		&parallelscommon.StepUploadVersion{
			Path: b.config.PrlctlVersionFile,
		},
		&parallelscommon.StepUploadParallelsTools{
			ParallelsToolsFlavor:    b.config.ParallelsToolsFlavor,
			ParallelsToolsGuestPath: b.config.ParallelsToolsGuestPath,
			ParallelsToolsMode:      b.config.ParallelsToolsMode,
			Ctx:                     b.config.ctx,
		},
		new(common.StepProvision),
		&parallelscommon.StepShutdown{
			Command: b.config.ShutdownCommand,
			Timeout: b.config.ShutdownTimeout,
		},
		&common.StepCleanupTempKeys{
			Comm: &b.config.SSHConfig.Comm,
		},
		&parallelscommon.StepPrlctl{
			Commands: b.config.PrlctlPost,
			Ctx:      b.config.ctx,
		},
		&parallelscommon.StepCompactDisk{
			Skip: b.config.SkipCompaction,
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

	return parallelscommon.NewArtifact(b.config.OutputDir)
}

// Cancel.
func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
