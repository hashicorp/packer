package jdcloud

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"boot_command",
			},
		},
	}, raws...)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Failed in decoding JSON->mapstructure")
	}

	errs := &packer.MultiError{}
	errs = packer.MultiErrorAppend(errs, b.config.JDCloudCredentialConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.JDCloudInstanceSpecConfig.Prepare(&b.config.ctx)...)
	if errs != nil && len(errs.Errors) != 0 {
		return nil, errs
	}

	packer.LogSecretFilter.Set(b.config.AccessKey, b.config.SecretKey)

	return nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {

	state := new(multistep.BasicStateBag)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("config", b.config)

	steps := []multistep.Step{

		&stepValidateParameters{
			InstanceSpecConfig: &b.config.JDCloudInstanceSpecConfig,
		},

		&stepConfigCredentials{
			InstanceSpecConfig: &b.config.JDCloudInstanceSpecConfig,
		},

		&stepCreateJDCloudInstance{
			InstanceSpecConfig: &b.config.JDCloudInstanceSpecConfig,
			CredentialConfig:   &b.config.JDCloudCredentialConfig,
		},

		&communicator.StepConnect{
			Config:    &b.config.JDCloudInstanceSpecConfig.Comm,
			SSHConfig: b.config.JDCloudInstanceSpecConfig.Comm.SSHConfigFunc(),
			Host:      instanceHost,
		},

		&common.StepProvision{},

		&stepStopJDCloudInstance{
			InstanceSpecConfig: &b.config.JDCloudInstanceSpecConfig,
		},

		&stepCreateJDCloudImage{
			InstanceSpecConfig: &b.config.JDCloudInstanceSpecConfig,
		},
	}

	b.runner = common.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
	b.runner.Run(ctx, state)

	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	artifact := &Artifact{
		ImageId:  b.config.ArtifactId,
		RegionID: b.config.RegionId,
	}
	return artifact, nil
}
