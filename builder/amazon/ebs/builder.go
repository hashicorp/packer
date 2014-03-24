// The amazonebs package contains a packer.Builder implementation that
// builds AMIs for Amazon EC2.
//
// In general, there are two types of AMIs that can be created: ebs-backed or
// instance-store. This builder _only_ builds ebs-backed images.
package ebs

import (
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	awscommon "github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"log"
)

// The unique ID for this builder
const BuilderId = "mitchellh.amazonebs"

type config struct {
	common.PackerConfig    `mapstructure:",squash"`
	awscommon.AccessConfig `mapstructure:",squash"`
	awscommon.AMIConfig    `mapstructure:",squash"`
	awscommon.BlockDevices `mapstructure:",squash"`
	awscommon.RunConfig    `mapstructure:",squash"`

	tpl *packer.ConfigTemplate
}

type Builder struct {
	config config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	md, err := common.DecodeConfig(&b.config, raws...)
	if err != nil {
		return nil, err
	}

	b.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return nil, err
	}
	b.config.tpl.UserVars = b.config.PackerUserVars
	b.config.tpl.Funcs(awscommon.TemplateFuncs)

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)
	errs = packer.MultiErrorAppend(errs, b.config.AccessConfig.Prepare(b.config.tpl)...)
	errs = packer.MultiErrorAppend(errs, b.config.AMIConfig.Prepare(b.config.tpl)...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(b.config.tpl)...)

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	log.Println(common.ScrubConfig(b.config, b.config.AccessKey, b.config.SecretKey))
	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	region, err := b.config.Region()
	if err != nil {
		return nil, err
	}

	auth, err := b.config.AccessConfig.Auth()
	if err != nil {
		return nil, err
	}

	ec2conn := ec2.New(auth, region)

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("ec2", ec2conn)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	steps := []multistep.Step{
		&awscommon.StepKeyPair{
			Debug:          b.config.PackerDebug,
			DebugKeyPath:   fmt.Sprintf("ec2_%s.pem", b.config.PackerBuildName),
			KeyPairName:    b.config.TemporaryKeyPairName,
			PrivateKeyFile: b.config.SSHPrivateKeyFile,
		},
		&awscommon.StepSecurityGroup{
			SecurityGroupIds: b.config.SecurityGroupIds,
			SSHPort:          b.config.SSHPort,
			VpcId:            b.config.VpcId,
		},
		&awscommon.StepRunSourceInstance{
			Debug:                    b.config.PackerDebug,
			ExpectedRootDevice:       "ebs",
			InstanceType:             b.config.InstanceType,
			UserData:                 b.config.UserData,
			UserDataFile:             b.config.UserDataFile,
			SourceAMI:                b.config.SourceAmi,
			IamInstanceProfile:       b.config.IamInstanceProfile,
			SubnetId:                 b.config.SubnetId,
			AssociatePublicIpAddress: b.config.AssociatePublicIpAddress,
			AvailabilityZone:         b.config.AvailabilityZone,
			BlockDevices:             b.config.BlockDevices,
			Tags:                     b.config.RunTags,
		},
		&common.StepConnectSSH{
			SSHAddress:     awscommon.SSHAddress(ec2conn, b.config.SSHPort),
			SSHConfig:      awscommon.SSHConfig(b.config.SSHUsername),
			SSHWaitTimeout: b.config.SSHTimeout(),
		},
		&common.StepProvision{},
		&stepStopInstance{},
		&stepCreateAMI{},
		&awscommon.StepAMIRegionCopy{
			Regions: b.config.AMIRegions,
		},
		&awscommon.StepModifyAMIAttributes{
			Description: b.config.AMIDescription,
			Users:       b.config.AMIUsers,
			Groups:      b.config.AMIGroups,
		},
		&awscommon.StepCreateTags{
			Tags: b.config.AMITags,
		},
	}

	// Run!
	if b.config.PackerDebug {
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.runner = &multistep.BasicRunner{Steps: steps}
	}

	b.runner.Run(state)

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
		Conn:           ec2conn,
	}

	return artifact, nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}
