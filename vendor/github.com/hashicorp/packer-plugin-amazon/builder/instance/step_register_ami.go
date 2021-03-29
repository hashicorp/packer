package instance

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	awscommon "github.com/hashicorp/packer-plugin-amazon/builder/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/random"
	confighelper "github.com/hashicorp/packer-plugin-sdk/template/config"
)

type StepRegisterAMI struct {
	PollingConfig            *awscommon.AWSPollingConfig
	EnableAMIENASupport      confighelper.Trilean
	EnableAMISriovNetSupport bool
	AMISkipBuildRegion       bool
}

func (s *StepRegisterAMI) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ec2conn := state.Get("ec2").(*ec2.EC2)
	manifestPath := state.Get("remote_manifest_path").(string)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Registering the AMI...")

	// Create the image
	amiName := config.AMIName
	state.Put("intermediary_image", false)
	if config.AMIEncryptBootVolume.True() || s.AMISkipBuildRegion {
		state.Put("intermediary_image", true)

		// From AWS SDK docs: You can encrypt a copy of an unencrypted snapshot,
		// but you cannot use it to create an unencrypted copy of an encrypted
		// snapshot. Your default CMK for EBS is used unless you specify a
		// non-default key using KmsKeyId.

		// If encrypt_boot is nil or true, we need to create a temporary image
		// so that in step_region_copy, we can copy it with the correct
		// encryption
		amiName = random.AlphaNum(7)
	}

	registerOpts := &ec2.RegisterImageInput{
		ImageLocation:       &manifestPath,
		Name:                aws.String(amiName),
		BlockDeviceMappings: config.AMIMappings.BuildEC2BlockDeviceMappings(),
	}

	if config.AMIVirtType != "" {
		registerOpts.VirtualizationType = aws.String(config.AMIVirtType)
	}

	if s.EnableAMISriovNetSupport {
		// Set SriovNetSupport to "simple". See http://goo.gl/icuXh5
		// As of February 2017, this applies to C3, C4, D2, I2, R3, and M4 (excluding m4.16xlarge)
		registerOpts.SriovNetSupport = aws.String("simple")
	}
	if s.EnableAMIENASupport.True() {
		// Set EnaSupport to true
		// As of February 2017, this applies to C5, I3, P2, R4, X1, and m4.16xlarge
		registerOpts.EnaSupport = aws.Bool(true)
	}

	registerResp, err := ec2conn.RegisterImage(registerOpts)
	if err != nil {
		state.Put("error", fmt.Errorf("Error registering AMI: %s", err))
		ui.Error(state.Get("error").(error).Error())
		return multistep.ActionHalt
	}

	// Set the AMI ID in the state
	ui.Say(fmt.Sprintf("AMI: %s", *registerResp.ImageId))
	amis := make(map[string]string)
	amis[*ec2conn.Config.Region] = *registerResp.ImageId
	state.Put("amis", amis)

	// Wait for the image to become ready
	ui.Say("Waiting for AMI to become ready...")
	if err := s.PollingConfig.WaitUntilAMIAvailable(ctx, ec2conn, *registerResp.ImageId); err != nil {
		err := fmt.Errorf("Error waiting for AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("snapshots", map[string][]string{})

	return multistep.ActionContinue
}

func (s *StepRegisterAMI) Cleanup(multistep.StateBag) {}
