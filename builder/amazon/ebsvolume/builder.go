// The ebsvolume package contains a packer.Builder implementation that
// builds EBS volumes for Amazon EC2 using an ephemeral instance,
package ebsvolume

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/errwrap"
	"github.com/mitchellh/multistep"
	awscommon "github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

const BuilderId = "mitchellh.amazon.ebsvolume"

type Config struct {
	common.PackerConfig    `mapstructure:",squash"`
	awscommon.AccessConfig `mapstructure:",squash"`
	awscommon.RunConfig    `mapstructure:",squash"`

	VolumeMappings        []BlockDevice `mapstructure:"ebs_volumes"`
	AMIEnhancedNetworking bool          `mapstructure:"enhanced_networking"`

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
	}, raws...)
	if err != nil {
		return nil, err
	}

	// Accumulate any errors
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, b.config.AccessConfig.Prepare(&b.config.ctx)...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(&b.config.ctx)...)

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	log.Println(common.ScrubConfig(b.config, b.config.AccessKey, b.config.SecretKey))
	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	config, err := b.config.Config()
	if err != nil {
		return nil, err
	}

	session, err := session.NewSession(config)
	if err != nil {
		return nil, errwrap.Wrapf("Error creating AWS Session: {{err}}", err)
	}

	ec2conn := ec2.New(session)

	// If the subnet is specified but not the AZ, try to determine the AZ automatically
	if b.config.SubnetId != "" && b.config.AvailabilityZone == "" {
		log.Printf("[INFO] Finding AZ for the given subnet '%s'", b.config.SubnetId)
		resp, err := ec2conn.DescribeSubnets(&ec2.DescribeSubnetsInput{SubnetIds: []*string{&b.config.SubnetId}})
		if err != nil {
			return nil, err
		}
		b.config.AvailabilityZone = *resp.Subnets[0].AvailabilityZone
		log.Printf("[INFO] AZ found: '%s'", b.config.AvailabilityZone)
	}

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("ec2", ec2conn)
	state.Put("hook", hook)
	state.Put("ui", ui)

	launchBlockDevices := commonBlockDevices(b.config.VolumeMappings)

	// Build the steps
	steps := []multistep.Step{
		&awscommon.StepSourceAMIInfo{
			SourceAmi:          b.config.SourceAmi,
			EnhancedNetworking: b.config.AMIEnhancedNetworking,
			AmiFilters:         b.config.SourceAmiFilter,
		},
		&awscommon.StepKeyPair{
			Debug:                b.config.PackerDebug,
			SSHAgentAuth:         b.config.Comm.SSHAgentAuth,
			DebugKeyPath:         fmt.Sprintf("ec2_%s.pem", b.config.PackerBuildName),
			KeyPairName:          b.config.SSHKeyPairName,
			TemporaryKeyPairName: b.config.TemporaryKeyPairName,
			PrivateKeyFile:       b.config.RunConfig.Comm.SSHPrivateKey,
		},
		&awscommon.StepSecurityGroup{
			SecurityGroupIds: b.config.SecurityGroupIds,
			CommConfig:       &b.config.RunConfig.Comm,
			VpcId:            b.config.VpcId,
		},
		&awscommon.StepRunSourceInstance{
			Debug:                    b.config.PackerDebug,
			ExpectedRootDevice:       "ebs",
			SpotPrice:                b.config.SpotPrice,
			SpotPriceProduct:         b.config.SpotPriceAutoProduct,
			InstanceType:             b.config.InstanceType,
			UserData:                 b.config.UserData,
			UserDataFile:             b.config.UserDataFile,
			SourceAMI:                b.config.SourceAmi,
			IamInstanceProfile:       b.config.IamInstanceProfile,
			SubnetId:                 b.config.SubnetId,
			AssociatePublicIpAddress: b.config.AssociatePublicIpAddress,
			EbsOptimized:             b.config.EbsOptimized,
			AvailabilityZone:         b.config.AvailabilityZone,
			BlockDevices:             launchBlockDevices,
			Tags:                     b.config.RunTags,
			InstanceInitiatedShutdownBehavior: b.config.InstanceInitiatedShutdownBehavior,
		},
		&stepTagEBSVolumes{
			VolumeMapping: b.config.VolumeMappings,
		},
		&awscommon.StepGetPassword{
			Debug:   b.config.PackerDebug,
			Comm:    &b.config.RunConfig.Comm,
			Timeout: b.config.WindowsPasswordTimeout,
		},
		&communicator.StepConnect{
			Config: &b.config.RunConfig.Comm,
			Host: awscommon.SSHHost(
				ec2conn,
				b.config.SSHPrivateIp),
			SSHConfig: awscommon.SSHConfig(
				b.config.RunConfig.Comm.SSHAgentAuth,
				b.config.RunConfig.Comm.SSHUsername,
				b.config.RunConfig.Comm.SSHPassword),
		},
		&common.StepProvision{},
		&awscommon.StepStopEBSBackedInstance{
			SpotPrice:           b.config.SpotPrice,
			DisableStopInstance: b.config.DisableStopInstance,
		},
		&awscommon.StepModifyEBSBackedInstance{
			EnableEnhancedNetworking: b.config.AMIEnhancedNetworking,
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
