package triton

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

const (
	BuilderId = "joyent.triton"
)

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	errs := &multierror.Error{}

	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
	}, raws...)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	errs = multierror.Append(errs, b.config.AccessConfig.Prepare(&b.config.ctx)...)
	errs = multierror.Append(errs, b.config.SourceMachineConfig.Prepare(&b.config.ctx)...)
	errs = multierror.Append(errs, b.config.Comm.Prepare(&b.config.ctx)...)
	errs = multierror.Append(errs, b.config.TargetImageConfig.Prepare(&b.config.ctx)...)

	// If we are using an SSH agent to sign requests, and no private key has been
	// specified for SSH, use the agent for connecting for provisioning.
	if b.config.AccessConfig.KeyMaterial == "" && b.config.Comm.SSHPrivateKeyFile == "" {
		b.config.Comm.SSHAgentAuth = true
	}

	return nil, errs.ErrorOrNil()
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	config := b.config

	driver, err := NewDriverTriton(ui, config)
	if err != nil {
		return nil, err
	}

	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("debug", b.config.PackerDebug)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	steps := []multistep.Step{
		&StepCreateSourceMachine{},
		&communicator.StepConnect{
			Config:    &config.Comm,
			Host:      commHost(b.config.Comm.SSHHost),
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		},
		&common.StepProvision{},
		&common.StepCleanupTempKeys{
			Comm: &config.Comm,
		},
		&StepStopMachine{},
		&StepCreateImageFromMachine{},
		&StepDeleteMachine{},
	}

	b.runner = common.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If there is no image, just return
	if _, ok := state.GetOk("image"); !ok {
		return nil, nil
	}

	artifact := &Artifact{
		ImageID:        state.Get("image").(string),
		BuilderIDValue: BuilderId,
		Driver:         driver,
	}

	return artifact, nil
}

// Cancel cancels a possibly running Builder. This should block until
// the builder actually cancels and cleans up after itself.
