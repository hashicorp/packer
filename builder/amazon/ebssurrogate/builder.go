// The ebssurrogate package contains a packer.Builder implementation that
// builds a new EBS-backed AMI using an ephemeral instance.
package ebssurrogate

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

const BuilderId = "mitchellh.amazon.ebssurrogate"

type Config struct {
	common.PackerConfig    `mapstructure:",squash"`
	awscommon.AccessConfig `mapstructure:",squash"`
	awscommon.RunConfig    `mapstructure:",squash"`
	awscommon.BlockDevices `mapstructure:",squash"`
	awscommon.AMIConfig    `mapstructure:",squash"`
	// A block device mapping
    // describing the root device of the AMI. This looks like the mappings in
    // ami_block_device_mapping, except with an additional field:
	RootDevice    RootBlockDevice  `mapstructure:"ami_root_device" required:"true"`
	VolumeRunTags awscommon.TagMap `mapstructure:"run_volume_tags"`
	// what architecture to use when registering the
    // final AMI; valid options are "x86_64" or "arm64". Defaults to "x86_64".
	Architecture  string           `mapstructure:"ami_architecture" required:"false"`

	ctx interpolate.Context
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	b.config.ctx.Funcs = awscommon.TemplateFuncs
	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"ami_description",
				"run_tags",
				"run_volume_tags",
				"snapshot_tags",
				"spot_tags",
				"tags",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	if b.config.PackerConfig.PackerForce {
		b.config.AMIForceDeregister = true
	}

	// Accumulate any errors
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, b.config.AccessConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs,
		b.config.AMIConfig.Prepare(&b.config.AccessConfig, &b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.BlockDevices.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.RootDevice.Prepare(&b.config.ctx)...)

	if b.config.AMIVirtType == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("ami_virtualization_type is required."))
	}

	foundRootVolume := false
	for _, launchDevice := range b.config.BlockDevices.LaunchMappings {
		if launchDevice.DeviceName == b.config.RootDevice.SourceDeviceName {
			foundRootVolume = true
			if launchDevice.OmitFromArtifact {
				errs = packer.MultiErrorAppend(errs, fmt.Errorf("You cannot set \"omit_from_artifact\": \"true\" for the root volume."))
			}
		}
	}

	if !foundRootVolume {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("no volume with name '%s' is found", b.config.RootDevice.SourceDeviceName))
	}

	if b.config.IsSpotInstance() && ((b.config.AMIENASupport != nil && *b.config.AMIENASupport) || b.config.AMISriovNetSupport) {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("Spot instances do not support modification, which is required "+
				"when either `ena_support` or `sriov_support` are set. Please ensure "+
				"you use an AMI that already has either SR-IOV or ENA enabled."))
	}

	if b.config.Architecture == "" {
		b.config.Architecture = "x86_64"
	}
	valid := false
	for _, validArch := range []string{"x86_64", "arm64"} {
		if validArch == b.config.Architecture {
			valid = true
			break
		}
	}
	if !valid {
		errs = packer.MultiErrorAppend(errs, errors.New(`The only valid ami_architecture values are "x86_64" and "arm64"`))
	}
	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	packer.LogSecretFilter.Set(b.config.AccessKey, b.config.SecretKey, b.config.Token)
	return nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	session, err := b.config.Session()
	if err != nil {
		return nil, err
	}
	ec2conn := ec2.New(session)

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("access_config", &b.config.AccessConfig)
	state.Put("ami_config", &b.config.AMIConfig)
	state.Put("ec2", ec2conn)
	state.Put("awsSession", session)
	state.Put("hook", hook)
	state.Put("ui", ui)

	var instanceStep multistep.Step

	if b.config.IsSpotInstance() {
		instanceStep = &awscommon.StepRunSpotInstance{
			AssociatePublicIpAddress:          b.config.AssociatePublicIpAddress,
			BlockDevices:                      b.config.BlockDevices,
			BlockDurationMinutes:              b.config.BlockDurationMinutes,
			Ctx:                               b.config.ctx,
			Comm:                              &b.config.RunConfig.Comm,
			Debug:                             b.config.PackerDebug,
			EbsOptimized:                      b.config.EbsOptimized,
			ExpectedRootDevice:                "ebs",
			IamInstanceProfile:                b.config.IamInstanceProfile,
			InstanceInitiatedShutdownBehavior: b.config.InstanceInitiatedShutdownBehavior,
			InstanceType:                      b.config.InstanceType,
			SourceAMI:                         b.config.SourceAmi,
			SpotPrice:                         b.config.SpotPrice,
			SpotPriceProduct:                  b.config.SpotPriceAutoProduct,
			SpotInstanceTypes:                 b.config.SpotInstanceTypes,
			SpotTags:                          b.config.SpotTags,
			Tags:                              b.config.RunTags,
			UserData:                          b.config.UserData,
			UserDataFile:                      b.config.UserDataFile,
			VolumeTags:                        b.config.VolumeRunTags,
		}
	} else {
		instanceStep = &awscommon.StepRunSourceInstance{
			AssociatePublicIpAddress:          b.config.AssociatePublicIpAddress,
			BlockDevices:                      b.config.BlockDevices,
			Comm:                              &b.config.RunConfig.Comm,
			Ctx:                               b.config.ctx,
			Debug:                             b.config.PackerDebug,
			EbsOptimized:                      b.config.EbsOptimized,
			EnableT2Unlimited:                 b.config.EnableT2Unlimited,
			ExpectedRootDevice:                "ebs",
			IamInstanceProfile:                b.config.IamInstanceProfile,
			InstanceInitiatedShutdownBehavior: b.config.InstanceInitiatedShutdownBehavior,
			InstanceType:                      b.config.InstanceType,
			IsRestricted:                      b.config.IsChinaCloud() || b.config.IsGovCloud(),
			SourceAMI:                         b.config.SourceAmi,
			Tags:                              b.config.RunTags,
			UserData:                          b.config.UserData,
			UserDataFile:                      b.config.UserDataFile,
			VolumeTags:                        b.config.VolumeRunTags,
		}
	}

	amiDevices := b.config.BuildAMIDevices()
	launchDevices := b.config.BuildLaunchDevices()

	// Build the steps
	steps := []multistep.Step{
		&awscommon.StepPreValidate{
			DestAmiName:     b.config.AMIName,
			ForceDeregister: b.config.AMIForceDeregister,
		},
		&awscommon.StepSourceAMIInfo{
			SourceAmi:                b.config.SourceAmi,
			EnableAMISriovNetSupport: b.config.AMISriovNetSupport,
			EnableAMIENASupport:      b.config.AMIENASupport,
			AmiFilters:               b.config.SourceAmiFilter,
			AMIVirtType:              b.config.AMIVirtType,
		},
		&awscommon.StepNetworkInfo{
			VpcId:               b.config.VpcId,
			VpcFilter:           b.config.VpcFilter,
			SecurityGroupIds:    b.config.SecurityGroupIds,
			SecurityGroupFilter: b.config.SecurityGroupFilter,
			SubnetId:            b.config.SubnetId,
			SubnetFilter:        b.config.SubnetFilter,
			AvailabilityZone:    b.config.AvailabilityZone,
		},
		&awscommon.StepKeyPair{
			Debug:        b.config.PackerDebug,
			Comm:         &b.config.RunConfig.Comm,
			DebugKeyPath: fmt.Sprintf("ec2_%s.pem", b.config.PackerBuildName),
		},
		&awscommon.StepSecurityGroup{
			SecurityGroupFilter:    b.config.SecurityGroupFilter,
			SecurityGroupIds:       b.config.SecurityGroupIds,
			CommConfig:             &b.config.RunConfig.Comm,
			TemporarySGSourceCidrs: b.config.TemporarySGSourceCidrs,
		},
		&awscommon.StepCleanupVolumes{
			BlockDevices: b.config.BlockDevices,
		},
		instanceStep,
		&awscommon.StepGetPassword{
			Debug:     b.config.PackerDebug,
			Comm:      &b.config.RunConfig.Comm,
			Timeout:   b.config.WindowsPasswordTimeout,
			BuildName: b.config.PackerBuildName,
		},
		&communicator.StepConnect{
			Config: &b.config.RunConfig.Comm,
			Host: awscommon.SSHHost(
				ec2conn,
				b.config.Comm.SSHInterface),
			SSHConfig: b.config.RunConfig.Comm.SSHConfigFunc(),
		},
		&common.StepProvision{},
		&common.StepCleanupTempKeys{
			Comm: &b.config.RunConfig.Comm,
		},
		&awscommon.StepStopEBSBackedInstance{
			Skip:                b.config.IsSpotInstance(),
			DisableStopInstance: b.config.DisableStopInstance,
		},
		&awscommon.StepModifyEBSBackedInstance{
			EnableAMISriovNetSupport: b.config.AMISriovNetSupport,
			EnableAMIENASupport:      b.config.AMIENASupport,
		},
		&StepSnapshotVolumes{
			LaunchDevices:   launchDevices,
			SnapshotOmitMap: b.config.GetOmissions(),
		},
		&awscommon.StepDeregisterAMI{
			AccessConfig:        &b.config.AccessConfig,
			ForceDeregister:     b.config.AMIForceDeregister,
			ForceDeleteSnapshot: b.config.AMIForceDeleteSnapshot,
			AMIName:             b.config.AMIName,
			Regions:             b.config.AMIRegions,
		},
		&StepRegisterAMI{
			RootDevice:               b.config.RootDevice,
			AMIDevices:               amiDevices,
			LaunchDevices:            launchDevices,
			EnableAMISriovNetSupport: b.config.AMISriovNetSupport,
			EnableAMIENASupport:      b.config.AMIENASupport,
			Architecture:             b.config.Architecture,
			LaunchOmitMap:            b.config.GetOmissions(),
		},
		&awscommon.StepAMIRegionCopy{
			AccessConfig:      &b.config.AccessConfig,
			Regions:           b.config.AMIRegions,
			AMIKmsKeyId:       b.config.AMIKmsKeyId,
			RegionKeyIds:      b.config.AMIRegionKMSKeyIDs,
			EncryptBootVolume: b.config.AMIEncryptBootVolume,
			Name:              b.config.AMIName,
			OriginalRegion:    *ec2conn.Config.Region,
		},
		&awscommon.StepModifyAMIAttributes{
			Description:    b.config.AMIDescription,
			Users:          b.config.AMIUsers,
			Groups:         b.config.AMIGroups,
			ProductCodes:   b.config.AMIProductCodes,
			SnapshotUsers:  b.config.SnapshotUsers,
			SnapshotGroups: b.config.SnapshotGroups,
			Ctx:            b.config.ctx,
		},
		&awscommon.StepCreateTags{
			Tags:         b.config.AMITags,
			SnapshotTags: b.config.SnapshotTags,
			Ctx:          b.config.ctx,
		},
	}

	// Run!
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	if amis, ok := state.GetOk("amis"); ok {
		// Build the artifact and return it
		artifact := &awscommon.Artifact{
			Amis:           amis.(map[string]string),
			BuilderIdValue: BuilderId,
			Session:        session,
		}

		return artifact, nil
	}

	return nil, nil
}
