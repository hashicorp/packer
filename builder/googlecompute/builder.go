// The googlecompute package contains a packer.Builder implementation that
// builds images for Google Compute Engine.
package googlecompute

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// The unique ID for this builder.
const BuilderId = "packer.googlecompute"

// Builder represents a Packer Builder.
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

// Run executes a googlecompute Packer build and returns a packer.Artifact
// representing a GCE machine image.
func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	driver, err := NewDriverGCE(
		ui, b.config.ProjectId, b.config.Account)
	if err != nil {
		return nil, err
	}

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps.
	steps := []multistep.Step{
		new(StepCheckExistingImage),
		&StepCreateSSHKey{
			Debug:        b.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("gce_%s.pem", b.config.PackerBuildName),
		},
		&StepCreateInstance{
			Debug: b.config.PackerDebug,
		},
		&StepCreateWindowsPassword{
			Debug:        b.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("gce_windows_%s.pem", b.config.PackerBuildName),
		},
		&StepInstanceInfo{
			Debug: b.config.PackerDebug,
		},
		&communicator.StepConnect{
			Config:      &b.config.Comm,
			Host:        communicator.CommHost(b.config.Comm.SSHHost, "instance_ip"),
			SSHConfig:   b.config.Comm.SSHConfigFunc(),
			WinRMConfig: winrmConfig,
		},
		new(common.StepProvision),
		&common.StepCleanupTempKeys{
			Comm: &b.config.Comm,
		},
	}
	if _, exists := b.config.Metadata[StartupScriptKey]; exists || b.config.StartupScriptFile != "" {
		steps = append(steps, new(StepWaitStartupScript))
	}
	steps = append(steps, new(StepTeardownInstance), new(StepCreateImage))

	// Run the steps.
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// Report any errors.
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}
	if _, ok := state.GetOk("image"); !ok {
		log.Println("Failed to find image in state. Bug?")
		return nil, nil
	}

	artifact := &Artifact{
		image:  state.Get("image").(*Image),
		driver: driver,
		config: b.config,
	}
	return artifact, nil
}

// Cancel.
