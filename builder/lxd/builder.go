package lxd

import (
	"context"
	"os/exec"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/builder/lxd/api"
	"github.com/hashicorp/packer/builder/lxd/cli"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

// The unique ID for this builder
const BuilderId = "lxd"

type wrappedCommandTemplate struct {
	Command string
}

type lxdClient interface {
	DeleteImage(string) error
	LaunchContainer(string, string, string, map[string]string) error
	PublishContainer(string, string, map[string]string) (string, error)
	StopContainer(string) error
	DeleteContainer(string) error
	ExecuteContainer(string, string, func(string) (string, error)) (*exec.Cmd, error)
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, []string, error) {
	errs := b.config.Prepare(raws...)
	if errs != nil {
		return nil, nil, errs
	}

	return nil, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	wrappedCommand := func(command string) (string, error) {
		b.config.ctx.Data = &wrappedCommandTemplate{Command: command}
		return interpolate.Render(b.config.CommandWrapper, &b.config.ctx)
	}

	var client lxdClient
	var err error
	client = &cli_client.LXDClient{}
	if b.config.LXDClient == "api" {
		client, err = api_client.NewLXDClient("")
	}
	if err != nil {
		return nil, err
	}

	steps := []multistep.Step{
		&stepLxdLaunch{client: client},
		&StepProvision{},
		&stepPublish{client: client},
	}

	// Setup the state bag
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("wrappedCommand", CommandWrapper(wrappedCommand))

	// Run
	b.runner = common.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	artifact := &Artifact{
		id:        state.Get("imageFingerprint").(string),
		StateData: map[string]interface{}{"generated_data": state.Get("generated_data")},
		client:    client,
	}

	return artifact, nil
}
