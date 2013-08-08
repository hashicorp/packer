// The amazonebs package contains a packer.Builder implementation that
// builds AMIs for Amazon EC2.
//
// In general, there are two types of AMIs that can be created: ebs-backed or
// instance-store. This builder _only_ builds ebs-backed images.
package ebs

import (
	"errors"
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
	awscommon.RunConfig    `mapstructure:",squash"`

	// Configuration of the resulting AMI
	AMIName string `mapstructure:"ami_name"`

	// Tags for the AMI
	Tags map[string]string

	tpl *common.Template
}

type Builder struct {
	config config
	runner multistep.Runner
}

func (b *Builder) Prepare(raws ...interface{}) error {
	md, err := common.DecodeConfig(&b.config, raws...)
	if err != nil {
		return err
	}

	b.config.tpl, err = common.NewTemplate()
	if err != nil {
		return err
	}

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)
	errs = packer.MultiErrorAppend(errs, b.config.AccessConfig.Prepare(b.config.tpl)...)
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(b.config.tpl)...)

	// Accumulate any errors
	if b.config.AMIName == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("ami_name must be specified"))
	} else if b.config.AMIName, err = b.config.tpl.Process(b.config.AMIName, nil); err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Error processing ami_name: %s", err))
	}

	newTags := make(map[string]string)
	for k, v := range b.config.Tags {
		k, err = b.config.tpl.Process(k, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Error processing tag key %s: %s", k, err))
			continue
		}

		v, err = b.config.tpl.Process(v, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("Error processing tag value '%s': %s", v, err))
			continue
		}

		newTags[k] = v
	}

	b.config.Tags = newTags

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	log.Printf("Config: %+v", b.config)
	return nil
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
	state := make(map[string]interface{})
	state["config"] = b.config
	state["ec2"] = ec2conn
	state["hook"] = hook
	state["ui"] = ui

	// Build the steps
	steps := []multistep.Step{
		&awscommon.StepKeyPair{},
		&awscommon.StepSecurityGroup{
			SecurityGroupId: b.config.SecurityGroupId,
			SSHPort:         b.config.SSHPort,
			VpcId:           b.config.VpcId,
		},
		&awscommon.StepRunSourceInstance{
			ExpectedRootDevice: "ebs",
			InstanceType:       b.config.InstanceType,
			SourceAMI:          b.config.SourceAmi,
			IamInstanceProfile: b.config.IamInstanceProfile,
			SubnetId:           b.config.SubnetId,
		},
		&common.StepConnectSSH{
			SSHAddress:     awscommon.SSHAddress(ec2conn, b.config.SSHPort),
			SSHConfig:      awscommon.SSHConfig(b.config.SSHUsername),
			SSHWaitTimeout: b.config.SSHTimeout(),
		},
		&common.StepProvision{},
		&stepStopInstance{},
		&stepCreateAMI{},
		&awscommon.StepCreateTags{Tags: b.config.Tags},
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
	if rawErr, ok := state["error"]; ok {
		return nil, rawErr.(error)
	}

	// If there are no AMIs, then just return
	if _, ok := state["amis"]; !ok {
		return nil, nil
	}

	// Build the artifact and return it
	artifact := &awscommon.Artifact{
		Amis:           state["amis"].(map[string]string),
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
