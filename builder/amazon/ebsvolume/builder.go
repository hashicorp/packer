// The ebsvolume package contains a packer.Builder implementation that
// builds EBS volumes for Amazon EC2 using an ephemeral instance,
package ebsvolume

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

const BuilderId = "mitchellh.amazon.ebsvolume"

type Config struct {
	common.PackerConfig    `mapstructure:",squash"`
	awscommon.AccessConfig `mapstructure:",squash"`
	awscommon.RunConfig    `mapstructure:",squash"`

	VolumeMappings     []BlockDevice `mapstructure:"ebs_volumes"`
	AMIENASupport      bool          `mapstructure:"ena_support"`
	AMISriovNetSupport bool          `mapstructure:"sriov_support"`

	launchBlockDevices awscommon.BlockDevices
	ctx                interpolate.Context
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
				"run_tags",
				"spot_tags",
				"ebs_volumes",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	// Accumulate any errors
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, b.config.AccessConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.launchBlockDevices.Prepare(&b.config.ctx)...)

	for _, d := range b.config.VolumeMappings {
		if err := d.Prepare(&b.config.ctx); err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("AMIMapping: %s", err.Error()))
		}
	}

	b.config.launchBlockDevices, err = commonBlockDevices(b.config.VolumeMappings, &b.config.ctx)
	if err != nil {
		errs = packer.MultiErrorAppend(errs, err)
	}

	if b.config.IsSpotInstance() && (b.config.AMIENASupport || b.config.AMISriovNetSupport) {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("Spot instances do not support modification, which is required "+
				"when either `ena_support` or `sriov_support` are set. Please ensure "+
				"you use an AMI that already has either SR-IOV or ENA enabled."))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	packer.LogSecretFilter.Set(b.config.AccessKey, b.config.SecretKey, b.config.Token)
	log.Println(b.config)
	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	session, err := b.config.Session()
	if err != nil {
		return nil, err
	}
	ec2conn := ec2.New(session)

	// If the subnet is specified but not the VpcId or AZ, try to determine them automatically
	if b.config.SubnetId != "" && (b.config.AvailabilityZone == "" || b.config.VpcId == "") {
		log.Printf("[INFO] Finding AZ and VpcId for the given subnet '%s'", b.config.SubnetId)
		resp, err := ec2conn.DescribeSubnets(&ec2.DescribeSubnetsInput{SubnetIds: []*string{&b.config.SubnetId}})
		if err != nil {
			return nil, err
		}
		if b.config.AvailabilityZone == "" {
			b.config.AvailabilityZone = *resp.Subnets[0].AvailabilityZone
			log.Printf("[INFO] AvailabilityZone found: '%s'", b.config.AvailabilityZone)
		}
		if b.config.VpcId == "" {
			b.config.VpcId = *resp.Subnets[0].VpcId
			log.Printf("[INFO] VpcId found: '%s'", b.config.VpcId)
		}
	}

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("ec2", ec2conn)
	state.Put("hook", hook)
	state.Put("ui", ui)

	var instanceStep multistep.Step

	if b.config.IsSpotInstance() {
		instanceStep = &awscommon.StepRunSpotInstance{
			AssociatePublicIpAddress:          b.config.AssociatePublicIpAddress,
			AvailabilityZone:                  b.config.AvailabilityZone,
			BlockDevices:                      b.config.launchBlockDevices,
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
			SpotTags:                          b.config.SpotTags,
			SubnetId:                          b.config.SubnetId,
			Tags:                              b.config.RunTags,
			UserData:                          b.config.UserData,
			UserDataFile:                      b.config.UserDataFile,
		}
	} else {
		instanceStep = &awscommon.StepRunSourceInstance{
			AssociatePublicIpAddress:          b.config.AssociatePublicIpAddress,
			AvailabilityZone:                  b.config.AvailabilityZone,
			BlockDevices:                      b.config.launchBlockDevices,
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
			SubnetId:                          b.config.SubnetId,
			Tags:                              b.config.RunTags,
			UserData:                          b.config.UserData,
			UserDataFile:                      b.config.UserDataFile,
		}
	}

	// Build the steps
	steps := []multistep.Step{
		&awscommon.StepSourceAMIInfo{
			SourceAmi:                b.config.SourceAmi,
			EnableAMISriovNetSupport: b.config.AMISriovNetSupport,
			EnableAMIENASupport:      b.config.AMIENASupport,
			AmiFilters:               b.config.SourceAmiFilter,
		},
		&awscommon.StepKeyPair{
			Debug:        b.config.PackerDebug,
			Comm:         &b.config.RunConfig.Comm,
			DebugKeyPath: fmt.Sprintf("ec2_%s.pem", b.config.PackerBuildName),
		},
		&awscommon.StepSecurityGroup{
			SecurityGroupIds:      b.config.SecurityGroupIds,
			CommConfig:            &b.config.RunConfig.Comm,
			VpcId:                 b.config.VpcId,
			TemporarySGSourceCidr: b.config.TemporarySGSourceCidr,
		},
		instanceStep,
		&stepTagEBSVolumes{
			VolumeMapping: b.config.VolumeMappings,
			Ctx:           b.config.ctx,
		},
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
	}

	// Run!
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// Build the artifact and return it
	artifact := &Artifact{
		Volumes:        state.Get("ebsvolumes").(EbsVolumes),
		BuilderIdValue: BuilderId,
		Conn:           ec2conn,
	}
	ui.Say(fmt.Sprintf("Created Volumes: %s", artifact))
	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
