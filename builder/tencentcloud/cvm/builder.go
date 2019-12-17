//go:generate mapstructure-to-hcl2 -type Config

package cvm

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

const BuilderId = "tencent.cloud"

type Config struct {
	common.PackerConfig      `mapstructure:",squash"`
	TencentCloudAccessConfig `mapstructure:",squash"`
	TencentCloudImageConfig  `mapstructure:",squash"`
	TencentCloudRunConfig    `mapstructure:",squash"`

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
	errs = packer.MultiErrorAppend(errs, b.config.TencentCloudAccessConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.TencentCloudImageConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.TencentCloudRunConfig.Prepare(&b.config.ctx)...)
	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	packer.LogSecretFilter.Set(b.config.SecretId, b.config.SecretKey)

	return nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	cvmClient, vpcClient, err := b.config.Client()
	if err != nil {
		return nil, err
	}

	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("cvm_client", cvmClient)
	state.Put("vpc_client", vpcClient)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	var steps []multistep.Step
	steps = []multistep.Step{
		&stepPreValidate{},
		&stepCheckSourceImage{
			b.config.SourceImageId,
		},
		&stepConfigKeyPair{
			Debug:        b.config.PackerDebug,
			Comm:         &b.config.Comm,
			DebugKeyPath: fmt.Sprintf("cvm_%s.pem", b.config.PackerBuildName),
		},
		&stepConfigVPC{
			VpcId:     b.config.VpcId,
			CidrBlock: b.config.CidrBlock,
			VpcName:   b.config.VpcName,
		},
		&stepConfigSubnet{
			SubnetId:        b.config.SubnetId,
			SubnetCidrBlock: b.config.SubnectCidrBlock,
			SubnetName:      b.config.SubnetName,
			Zone:            b.config.Zone,
		},
		&stepConfigSecurityGroup{
			SecurityGroupId:   b.config.SecurityGroupId,
			SecurityGroupName: b.config.SecurityGroupName,
			Description:       "securitygroup for packer",
		},
		&stepRunInstance{
			InstanceType:             b.config.InstanceType,
			UserData:                 b.config.UserData,
			UserDataFile:             b.config.UserDataFile,
			ZoneId:                   b.config.Zone,
			InstanceName:             b.config.InstanceName,
			DiskType:                 b.config.DiskType,
			DiskSize:                 b.config.DiskSize,
			DataDisks:                b.config.DataDisks,
			HostName:                 b.config.HostName,
			InternetMaxBandwidthOut:  b.config.InternetMaxBandwidthOut,
			AssociatePublicIpAddress: b.config.AssociatePublicIpAddress,
			Tags:                     b.config.RunTags,
		},
		&communicator.StepConnect{
			Config:    &b.config.TencentCloudRunConfig.Comm,
			SSHConfig: b.config.TencentCloudRunConfig.Comm.SSHConfigFunc(),
			Host:      SSHHost(b.config.AssociatePublicIpAddress),
		},
		&common.StepProvision{},
		&common.StepCleanupTempKeys{
			Comm: &b.config.TencentCloudRunConfig.Comm,
		},
		// We need this step to detach keypair from instance, otherwise
		// it always fails to delete the key.
		&stepDetachTempKeyPair{},
		&stepCreateImage{},
		&stepShareImage{
			b.config.ImageShareAccounts,
		},
		&stepCopyImage{
			DesinationRegions: b.config.ImageCopyRegions,
			SourceRegion:      b.config.Region,
		},
	}

	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	if _, ok := state.GetOk("image"); !ok {
		return nil, nil
	}

	artifact := &Artifact{
		TencentCloudImages: state.Get("tencentcloudimages").(map[string]string),
		BuilderIdValue:     BuilderId,
		Client:             cvmClient,
	}

	return artifact, nil
}
