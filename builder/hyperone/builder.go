package hyperone

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	openapi "github.com/hyperonecom/h1-client-go"
)

const BuilderID = "hyperone.builder"

type Builder struct {
	config Config
	runner multistep.Runner
	client *openapi.APIClient
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	config, warnings, errs := NewConfig(raws...)
	if errs != nil {
		return warnings, errs
	}

	b.config = *config

	cfg := openapi.NewConfiguration()
	cfg.AddDefaultHeader("x-auth-token", b.config.Token)
	if b.config.Project != "" {
		cfg.AddDefaultHeader("x-project", b.config.Project)
	}

	if b.config.APIURL != "" {
		cfg.BasePath = b.config.APIURL
	}

	prefer := fmt.Sprintf("respond-async,wait=%d", int(b.config.StateTimeout.Seconds()))
	cfg.AddDefaultHeader("Prefer", prefer)

	b.client = openapi.NewAPIClient(cfg)

	return nil, nil
}

type wrappedCommandTemplate struct {
	Command string
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	wrappedCommand := func(command string) (string, error) {
		ctx := b.config.ctx
		ctx.Data = &wrappedCommandTemplate{Command: command}
		return interpolate.Render(b.config.ChrootCommandWrapper, &ctx)
	}

	state := &multistep.BasicStateBag{}
	state.Put("config", &b.config)
	state.Put("client", b.client)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("wrappedCommand", CommandWrapper(wrappedCommand))

	steps := []multistep.Step{
		&stepCreateSSHKey{},
		&stepCreateVM{},
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      getPublicIP,
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		},
	}

	if b.config.ChrootDisk {
		steps = append(steps,
			&stepPrepareDevice{},
			&stepPreMountCommands{},
			&stepMountChroot{},
			&stepPostMountCommands{},
			&stepMountExtra{},
			&stepCopyFiles{},
			&stepChrootProvision{},
			&stepStopVM{},
			&stepDetachDisk{},
			&stepCreateVMFromDisk{},
			&stepCreateImage{},
		)
	} else {
		steps = append(steps,
			&common.StepProvision{},
			&stepStopVM{},
			&stepCreateImage{},
		)
	}

	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	artifact := &Artifact{
		imageID:   state.Get("image_id").(string),
		imageName: state.Get("image_name").(string),
		client:    b.client,
	}

	return artifact, nil
}
