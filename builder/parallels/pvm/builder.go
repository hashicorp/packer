package pvm

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	parallelscommon "github.com/hashicorp/packer/builder/parallels/common"
)

// Builder implements packersdk.Builder and builds the actual Parallels
// images.
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

// Run executes a Packer build and returns a packersdk.Artifact representing
// a Parallels appliance.
func (b *Builder) Run(ctx context.Context, ui packersdk.Ui, hook packersdk.Hook) (packersdk.Artifact, error) {
	// Create the driver that we'll use to communicate with Parallels
	driver, err := parallelscommon.NewDriver()
	if err != nil {
		return nil, fmt.Errorf("Failed creating Parallels driver: %s", err)
	}

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
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
		&commonsteps.StepCreateFloppy{
			Files:       b.config.FloppyConfig.FloppyFiles,
			Directories: b.config.FloppyConfig.FloppyDirectories,
			Label:       b.config.FloppyConfig.FloppyLabel,
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
			Host:      parallelscommon.CommHost(b.config.SSHConfig.Comm.SSHHost),
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
		new(commonsteps.StepProvision),
		&parallelscommon.StepShutdown{
			Command: b.config.ShutdownCommand,
			Timeout: b.config.ShutdownTimeout,
		},
		&commonsteps.StepCleanupTempKeys{
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
	b.runner = commonsteps.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
	b.runner.Run(ctx, state)

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

	generatedData := map[string]interface{}{"generated_data": state.Get("generated_data")}
	return parallelscommon.NewArtifact(b.config.OutputDir, generatedData)
}

// Cancel.
