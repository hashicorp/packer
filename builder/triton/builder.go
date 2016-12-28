package triton

import (
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
)

const (
	BuilderId = "joyent.triton"
)

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	errs := &multierror.Error{}

	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
	}, raws...)
	if err != nil {
		errs = multierror.Append(errs, err)
	}

	// In Triton only the root user is setup in a VM.
	b.config.Comm.SSHUsername = "root"

	errs = multierror.Append(errs, b.config.AccessConfig.Prepare(&b.config.ctx)...)
	errs = multierror.Append(errs, b.config.SourceMachineConfig.Prepare(&b.config.ctx)...)
	errs = multierror.Append(errs, b.config.Comm.Prepare(&b.config.ctx)...)
	errs = multierror.Append(errs, b.config.TargetImageConfig.Prepare(&b.config.ctx)...)

	b.config.Comm.SSHPrivateKey = b.config.KeyMaterial

	return nil, errs.ErrorOrNil()
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	config := b.config

	driver, err := NewDriverTriton(ui, config)
	if err != nil {
		return nil, err
	}

	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("driver", driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	steps := []multistep.Step{
		&StepCreateSourceMachine{},
		&communicator.StepConnect{
			Config:    &config.Comm,
			Host:      commHost,
			SSHConfig: sshConfig,
		},
		&common.StepProvision{},
		&StepStopMachine{},
		&StepCreateImageFromMachine{},
		&StepDeleteMachine{},
	}

	if b.config.PackerDebug {
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.runner = &multistep.BasicRunner{Steps: steps}
	}

	b.runner.Run(state)

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
func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
