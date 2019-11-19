//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

// The amazonebs package contains a packer.Builder implementation that
// builds AMIs for Amazon EC2.
//
// In general, there are two types of AMIs that can be created: ebs-backed or
// instance-store. This builder _only_ builds ebs-backed images.
package ebs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

// The unique ID for this builder
const BuilderId = "mitchellh.amazonebs"

type Config struct {
	common.PackerConfig    `mapstructure:",squash"`
	awscommon.AccessConfig `mapstructure:",squash"`
	awscommon.AMIConfig    `mapstructure:",squash"`
	awscommon.RunConfig    `mapstructure:",squash"`
	// Add one or more block device mappings to the AMI. These will be attached
	// when booting a new instance from your AMI. To add a block device during
	// the Packer build see `launch_block_device_mappings` below. Your options
	// here may vary depending on the type of VM you use. See the
	// [BlockDevices](#block-devices-configuration) documentation for fields.
	AMIMappings awscommon.BlockDevices `mapstructure:"ami_block_device_mappings" required:"false"`
	// Add one or more block devices before the Packer build starts. If you add
	// instance store volumes or EBS volumes in addition to the root device
	// volume, the created AMI will contain block device mapping information
	// for those volumes. Amazon creates snapshots of the source instance's
	// root volume and any other EBS volumes described here. When you launch an
	// instance from this new AMI, the instance automatically launches with
	// these additional volumes, and will restore them from snapshots taken
	// from the source instance. See the
	// [BlockDevices](#block-devices-configuration) documentation for fields.
	LaunchMappings awscommon.BlockDevices `mapstructure:"launch_block_device_mappings" required:"false"`
	// Tags to apply to the volumes that are *launched* to create the AMI.
	// These tags are *not* applied to the resulting AMI unless they're
	// duplicated in `tags`. This is a [template
	// engine](/docs/templates/engine.html), see [Build template
	// data](#build-template-data) for more information.
	VolumeRunTags awscommon.TagMap `mapstructure:"run_volume_tags"`
	// If you set this flag, we'll add clauses to the
	// launch_block_device_mappings that make sure ephemeral drives don't show
	// up in the EC2 console. If you launched from the EC2 console, you'd get
	// this automatically, but the SDK does not provide this service.
	// This only applies when you are not running spot instances.
	NoEphemeral bool `mapstructure:"no_ephemeral" required:"false"`

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
				"spot_tags",
				"snapshot_tags",
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
	var warns []string

	errs = packer.MultiErrorAppend(errs, b.config.AccessConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs,
		b.config.AMIConfig.Prepare(&b.config.AccessConfig, &b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.AMIMappings.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.LaunchMappings.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(&b.config.ctx)...)

	if b.config.IsSpotInstance() && (b.config.AMIENASupport.True() || b.config.AMISriovNetSupport) {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("Spot instances do not support modification, which is required "+
				"when either `ena_support` or `sriov_support` are set. Please ensure "+
				"you use an AMI that already has either SR-IOV or ENA enabled."))
	}

	if b.config.RunConfig.SpotPriceAutoProduct != "" {
		warns = append(warns, "spot_price_auto_product is deprecated and no "+
			"longer necessary for Packer builds. In future versions of "+
			"Packer, inclusion of spot_price_auto_product will error your "+
			"builds. Please take a look at our current documentation to "+
			"understand how Packer requests Spot instances.")
	}

	if errs != nil && len(errs.Errors) > 0 {
		return warns, errs
	}

	packer.LogSecretFilter.Set(b.config.AccessKey, b.config.SecretKey, b.config.Token)
	return warns, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {

	session, err := b.config.Session()
	if err != nil {
		return nil, err
	}

	ec2conn := ec2.New(session)
	iam := iam.New(session)
	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("access_config", &b.config.AccessConfig)
	state.Put("ami_config", &b.config.AMIConfig)
	state.Put("ec2", ec2conn)
	state.Put("iam", iam)
	state.Put("awsSession", session)
	state.Put("hook", hook)
	state.Put("ui", ui)

	var instanceStep multistep.Step

	if b.config.IsSpotInstance() {
		instanceStep = &awscommon.StepRunSpotInstance{
			AssociatePublicIpAddress:          b.config.AssociatePublicIpAddress,
			LaunchMappings:                    b.config.LaunchMappings,
			BlockDurationMinutes:              b.config.BlockDurationMinutes,
			Ctx:                               b.config.ctx,
			Comm:                              &b.config.RunConfig.Comm,
			Debug:                             b.config.PackerDebug,
			EbsOptimized:                      b.config.EbsOptimized,
			ExpectedRootDevice:                "ebs",
			InstanceInitiatedShutdownBehavior: b.config.InstanceInitiatedShutdownBehavior,
			InstanceType:                      b.config.InstanceType,
			SourceAMI:                         b.config.SourceAmi,
			SpotPrice:                         b.config.SpotPrice,
			SpotTags:                          b.config.SpotTags,
			Tags:                              b.config.RunTags,
			SpotInstanceTypes:                 b.config.SpotInstanceTypes,
			UserData:                          b.config.UserData,
			UserDataFile:                      b.config.UserDataFile,
			VolumeTags:                        b.config.VolumeRunTags,
		}
	} else {
		instanceStep = &awscommon.StepRunSourceInstance{
			AssociatePublicIpAddress:          b.config.AssociatePublicIpAddress,
			LaunchMappings:                    b.config.LaunchMappings,
			Comm:                              &b.config.RunConfig.Comm,
			Ctx:                               b.config.ctx,
			Debug:                             b.config.PackerDebug,
			EbsOptimized:                      b.config.EbsOptimized,
			EnableT2Unlimited:                 b.config.EnableT2Unlimited,
			ExpectedRootDevice:                "ebs",
			InstanceInitiatedShutdownBehavior: b.config.InstanceInitiatedShutdownBehavior,
			InstanceType:                      b.config.InstanceType,
			IsRestricted:                      b.config.IsChinaCloud() || b.config.IsGovCloud(),
			SourceAMI:                         b.config.SourceAmi,
			Tags:                              b.config.RunTags,
			UserData:                          b.config.UserData,
			UserDataFile:                      b.config.UserDataFile,
			VolumeTags:                        b.config.VolumeRunTags,
			NoEphemeral:                       b.config.NoEphemeral,
		}
	}

	// Build the steps
	steps := []multistep.Step{
		&awscommon.StepPreValidate{
			DestAmiName:        b.config.AMIName,
			ForceDeregister:    b.config.AMIForceDeregister,
			AMISkipBuildRegion: b.config.AMISkipBuildRegion,
			VpcId:              b.config.VpcId,
			SubnetId:           b.config.SubnetId,
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
		&awscommon.StepIamInstanceProfile{
			IamInstanceProfile:                        b.config.IamInstanceProfile,
			TemporaryIamInstanceProfilePolicyDocument: b.config.TemporaryIamInstanceProfilePolicyDocument,
		},
		&awscommon.StepCleanupVolumes{
			LaunchMappings: b.config.LaunchMappings,
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
				b.config.SSHInterface,
				b.config.Comm.SSHHost,
			),
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
		&awscommon.StepDeregisterAMI{
			AccessConfig:        &b.config.AccessConfig,
			ForceDeregister:     b.config.AMIForceDeregister,
			ForceDeleteSnapshot: b.config.AMIForceDeleteSnapshot,
			AMIName:             b.config.AMIName,
			Regions:             b.config.AMIRegions,
		},
		&stepCreateAMI{
			AMISkipBuildRegion: b.config.AMISkipBuildRegion,
		},
		&awscommon.StepAMIRegionCopy{
			AccessConfig:       &b.config.AccessConfig,
			Regions:            b.config.AMIRegions,
			AMIKmsKeyId:        b.config.AMIKmsKeyId,
			RegionKeyIds:       b.config.AMIRegionKMSKeyIDs,
			EncryptBootVolume:  b.config.AMIEncryptBootVolume,
			Name:               b.config.AMIName,
			OriginalRegion:     *ec2conn.Config.Region,
			AMISkipBuildRegion: b.config.AMISkipBuildRegion,
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

	// If there are no AMIs, then just return
	if _, ok := state.GetOk("amis"); !ok {
		return nil, nil
	}

	// Build the artifact and return it
	artifact := &awscommon.Artifact{
		Amis:           state.Get("amis").(map[string]string),
		BuilderIdValue: BuilderId,
		Session:        session,
	}

	return artifact, nil
}
