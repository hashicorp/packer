//go:generate mapstructure-to-hcl2 -type Config

// The ucloud-uhost contains a packer.Builder implementation that
// builds uhost images for UCloud UHost instance.
package uhost

import (
	"context"

	"github.com/hashicorp/hcl/v2/hcldec"
	ucloudcommon "github.com/hashicorp/packer/builder/ucloud/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

// The unique ID for this builder
const BuilderId = "ucloud.uhost"

type Config struct {
	common.PackerConfig       `mapstructure:",squash"`
	ucloudcommon.AccessConfig `mapstructure:",squash"`
	ucloudcommon.ImageConfig  `mapstructure:",squash"`
	ucloudcommon.RunConfig    `mapstructure:",squash"`

	ctx interpolate.Context
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...)
	b.config.ctx.EnableEnv = true
	if err != nil {
		return nil, err
	}

	// Accumulate any errors
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, b.config.AccessConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.ImageConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(&b.config.ctx)...)

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	packer.LogSecretFilter.Set(b.config.PublicKey, b.config.PrivateKey)
	return nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {

	client, err := b.config.Client()
	if err != nil {
		return nil, err
	}

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("client", client)
	state.Put("hook", hook)
	state.Put("ui", ui)

	var steps []multistep.Step

	// Build the steps
	steps = []multistep.Step{
		&stepPreValidate{
			ProjectId:         b.config.ProjectId,
			Region:            b.config.Region,
			Zone:              b.config.Zone,
			ImageDestinations: b.config.ImageDestinations,
		},

		&stepCheckSourceImageId{
			SourceUHostImageId: b.config.SourceImageId,
		},

		&stepConfigVPC{
			VPCId: b.config.VPCId,
		},
		&stepConfigSubnet{
			SubnetId: b.config.SubnetId,
		},
		&stepConfigSecurityGroup{
			SecurityGroupId: b.config.SecurityGroupId,
		},

		&stepCreateInstance{
			InstanceType:  b.config.InstanceType,
			Region:        b.config.Region,
			Zone:          b.config.Zone,
			SourceImageId: b.config.SourceImageId,
			InstanceName:  b.config.InstanceName,
			BootDiskType:  b.config.BootDiskType,
			UsePrivateIp:  b.config.UseSSHPrivateIp,
		},
		&communicator.StepConnect{
			Config: &b.config.RunConfig.Comm,
			Host: ucloudcommon.SSHHost(
				b.config.UseSSHPrivateIp),
			SSHConfig: b.config.RunConfig.Comm.SSHConfigFunc(),
		},
		&common.StepProvision{},
		&stepStopInstance{},
		&stepCreateImage{},
		&stepCopyUCloudImage{
			ImageDestinations:     b.config.ImageDestinations,
			RegionId:              b.config.Region,
			ProjectId:             b.config.ProjectId,
			WaitImageReadyTimeout: b.config.WaitImageReadyTimeout,
		},
	}

	// Run!
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If there are no ucloud images, then just return
	if _, ok := state.GetOk("ucloud_images"); !ok {
		return nil, nil
	}

	// Build the artifact and return it
	artifact := &ucloudcommon.Artifact{
		UCloudImages:   state.Get("ucloud_images").(*ucloudcommon.ImageInfoSet),
		BuilderIdValue: BuilderId,
		Client:         client,
	}

	return artifact, nil
}
