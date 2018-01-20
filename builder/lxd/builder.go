package lxd

import (
	"log"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"log"
)

// The unique ID for this builder
const BuilderId = "lxd"

type wrappedCommandTemplate struct {
	Command string
}

type Builder struct {
	config *Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	c, errs := NewConfig(raws...)
	if errs != nil {
		return nil, errs
	}
	b.config = c

	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	wrappedCommand := func(command string) (string, error) {
		b.config.ctx.Data = &wrappedCommandTemplate{Command: command}
		return interpolate.Render(b.config.CommandWrapper, &b.config.ctx)
	}

	steps := []multistep.Step{
		&stepLxdLaunch{},
		&StepProvision{},
		&stepPublish{},
	}

	// Setup the state bag
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("cache", cache)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("wrappedCommand", CommandWrapper(wrappedCommand))

	// Run
	b.runner = common.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
	b.runner.Run(state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	artifact := &Artifact{
		id: state.Get("imageFingerprint").(string),
	}

	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
