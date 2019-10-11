package null

import (
	"context"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

const BuilderId = "fnoeding.null"

type Builder struct {
	config *Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	c, warnings, errs := NewConfig(raws...)
	if errs != nil {
		return warnings, errs
	}
	b.config = c

	return warnings, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	steps := []multistep.Step{}

	steps = append(steps,
		&communicator.StepConnect{
			Config:    &b.config.CommConfig,
			Host:      CommHost(b.config.CommConfig.Host()),
			SSHConfig: b.config.CommConfig.SSHConfigFunc(),
		},
	)

	steps = append(steps,
		new(common.StepProvision),
	)

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Run!
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// No errors, must've worked
	artifact := &NullArtifact{}
	return artifact, nil
}
