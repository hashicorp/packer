//go:generate mapstructure-to-hcl2 -type Config,AlicloudDiskDevice

// The alicloud  contains a packer.Builder implementation that
// builds ecs images for alicloud.
package ecs

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

// The unique ID for this builder
const BuilderId = "alibaba.alicloud"

type Config struct {
	common.PackerConfig  `mapstructure:",squash"`
	AlicloudAccessConfig `mapstructure:",squash"`
	AlicloudImageConfig  `mapstructure:",squash"`
	RunConfig            `mapstructure:",squash"`

	ctx interpolate.Context
}

type Builder struct {
	config Config
	runner multistep.Runner
}

type InstanceNetWork string

const (
	ALICLOUD_DEFAULT_SHORT_TIMEOUT = 180
	ALICLOUD_DEFAULT_TIMEOUT       = 1800
	ALICLOUD_DEFAULT_LONG_TIMEOUT  = 3600
)

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

	if b.config.PackerConfig.PackerForce {
		b.config.AlicloudImageForceDelete = true
		b.config.AlicloudImageForceDeleteSnapshots = true
	}

	// Accumulate any errors
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, b.config.AlicloudAccessConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.AlicloudImageConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(&b.config.ctx)...)

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	packer.LogSecretFilter.Set(b.config.AlicloudAccessKey, b.config.AlicloudSecretKey)
	return nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {

	client, err := b.config.Client()
	if err != nil {
		return nil, err
	}
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("client", client)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("networktype", b.chooseNetworkType())
	var steps []multistep.Step

	// Build the steps
	steps = []multistep.Step{
		&stepPreValidate{
			AlicloudDestImageName: b.config.AlicloudImageName,
			ForceDelete:           b.config.AlicloudImageForceDelete,
		},
		&stepCheckAlicloudSourceImage{
			SourceECSImageId: b.config.AlicloudSourceImage,
		},
		&stepConfigAlicloudKeyPair{
			Debug:        b.config.PackerDebug,
			Comm:         &b.config.Comm,
			DebugKeyPath: fmt.Sprintf("ecs_%s.pem", b.config.PackerBuildName),
			RegionId:     b.config.AlicloudRegion,
		},
	}
	if b.chooseNetworkType() == InstanceNetworkVpc {
		steps = append(steps,
			&stepConfigAlicloudVPC{
				VpcId:     b.config.VpcId,
				CidrBlock: b.config.CidrBlock,
				VpcName:   b.config.VpcName,
			},
			&stepConfigAlicloudVSwitch{
				VSwitchId:   b.config.VSwitchId,
				ZoneId:      b.config.ZoneId,
				CidrBlock:   b.config.CidrBlock,
				VSwitchName: b.config.VSwitchName,
			})
	}
	steps = append(steps,
		&stepConfigAlicloudSecurityGroup{
			SecurityGroupId:   b.config.SecurityGroupId,
			SecurityGroupName: b.config.SecurityGroupId,
			RegionId:          b.config.AlicloudRegion,
		},
		&stepCreateAlicloudInstance{
			IOOptimized:             b.config.IOOptimized,
			InstanceType:            b.config.InstanceType,
			UserData:                b.config.UserData,
			UserDataFile:            b.config.UserDataFile,
			RegionId:                b.config.AlicloudRegion,
			InternetChargeType:      b.config.InternetChargeType,
			InternetMaxBandwidthOut: b.config.InternetMaxBandwidthOut,
			InstanceName:            b.config.InstanceName,
			ZoneId:                  b.config.ZoneId,
		})
	if b.chooseNetworkType() == InstanceNetworkVpc {
		steps = append(steps, &stepConfigAlicloudEIP{
			AssociatePublicIpAddress: b.config.AssociatePublicIpAddress,
			RegionId:                 b.config.AlicloudRegion,
			InternetChargeType:       b.config.InternetChargeType,
			InternetMaxBandwidthOut:  b.config.InternetMaxBandwidthOut,
			SSHPrivateIp:             b.config.SSHPrivateIp,
		})
	} else {
		steps = append(steps, &stepConfigAlicloudPublicIP{
			RegionId:     b.config.AlicloudRegion,
			SSHPrivateIp: b.config.SSHPrivateIp,
		})
	}
	steps = append(steps,
		&stepAttachKeyPair{},
		&stepRunAlicloudInstance{},
		&communicator.StepConnect{
			Config: &b.config.RunConfig.Comm,
			Host: SSHHost(
				client,
				b.config.SSHPrivateIp),
			SSHConfig: b.config.RunConfig.Comm.SSHConfigFunc(),
		},
		&common.StepProvision{},
		&common.StepCleanupTempKeys{
			Comm: &b.config.RunConfig.Comm,
		},
		&stepStopAlicloudInstance{
			ForceStop:   b.config.ForceStopInstance,
			DisableStop: b.config.DisableStopInstance,
		},
		&stepDeleteAlicloudImageSnapshots{
			AlicloudImageForceDeleteSnapshots: b.config.AlicloudImageForceDeleteSnapshots,
			AlicloudImageForceDelete:          b.config.AlicloudImageForceDelete,
			AlicloudImageName:                 b.config.AlicloudImageName,
			AlicloudImageDestinationRegions:   b.config.AlicloudImageConfig.AlicloudImageDestinationRegions,
			AlicloudImageDestinationNames:     b.config.AlicloudImageConfig.AlicloudImageDestinationNames,
		})

	if b.config.AlicloudImageIgnoreDataDisks {
		steps = append(steps, &stepCreateAlicloudSnapshot{
			WaitSnapshotReadyTimeout: b.getSnapshotReadyTimeout(),
		})
	}

	steps = append(steps,
		&stepCreateAlicloudImage{
			AlicloudImageIgnoreDataDisks: b.config.AlicloudImageIgnoreDataDisks,
			WaitSnapshotReadyTimeout:     b.getSnapshotReadyTimeout(),
		},
		&stepCreateTags{
			Tags: b.config.AlicloudImageTags,
		},
		&stepRegionCopyAlicloudImage{
			AlicloudImageDestinationRegions: b.config.AlicloudImageDestinationRegions,
			AlicloudImageDestinationNames:   b.config.AlicloudImageDestinationNames,
			RegionId:                        b.config.AlicloudRegion,
		},
		&stepShareAlicloudImage{
			AlicloudImageShareAccounts:   b.config.AlicloudImageShareAccounts,
			AlicloudImageUNShareAccounts: b.config.AlicloudImageUNShareAccounts,
			RegionId:                     b.config.AlicloudRegion,
		})

	// Run!
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If there are no ECS images, then just return
	if _, ok := state.GetOk("alicloudimages"); !ok {
		return nil, nil
	}

	// Build the artifact and return it
	artifact := &Artifact{
		AlicloudImages: state.Get("alicloudimages").(map[string]string),
		BuilderIdValue: BuilderId,
		Client:         client,
	}

	return artifact, nil
}

func (b *Builder) chooseNetworkType() InstanceNetWork {
	if b.isVpcNetRequired() {
		return InstanceNetworkVpc
	} else {
		return InstanceNetworkClassic
	}
}

func (b *Builder) isVpcNetRequired() bool {
	// UserData and KeyPair only works in VPC
	return b.isVpcSpecified() || b.isUserDataNeeded() || b.isKeyPairNeeded()
}

func (b *Builder) isVpcSpecified() bool {
	return b.config.VpcId != "" || b.config.VSwitchId != ""
}

func (b *Builder) isUserDataNeeded() bool {
	// Public key setup requires userdata
	if b.config.RunConfig.Comm.SSHPrivateKeyFile != "" {
		return true
	}

	return b.config.UserData != "" || b.config.UserDataFile != ""
}

func (b *Builder) isKeyPairNeeded() bool {
	return b.config.Comm.SSHKeyPairName != "" || b.config.Comm.SSHTemporaryKeyPairName != ""
}

func (b *Builder) getSnapshotReadyTimeout() int {
	if b.config.WaitSnapshotReadyTimeout > 0 {
		return b.config.WaitSnapshotReadyTimeout
	}

	return ALICLOUD_DEFAULT_LONG_TIMEOUT
}
