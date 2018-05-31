package instance

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepRegisterAMI struct {
	EnableAMIENASupport      bool
	EnableAMISriovNetSupport bool
}

func (s *StepRegisterAMI) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ec2conn := state.Get("ec2").(*ec2.EC2)
	manifestPath := state.Get("remote_manifest_path").(string)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Registering the AMI...")
	registerOpts := &ec2.RegisterImageInput{
		ImageLocation:       &manifestPath,
		Name:                aws.String(config.AMIName),
		BlockDeviceMappings: config.BlockDevices.BuildAMIDevices(),
	}

	if config.AMIVirtType != "" {
		registerOpts.VirtualizationType = aws.String(config.AMIVirtType)
	}

	if s.EnableAMISriovNetSupport {
		// Set SriovNetSupport to "simple". See http://goo.gl/icuXh5
		// As of February 2017, this applies to C3, C4, D2, I2, R3, and M4 (excluding m4.16xlarge)
		registerOpts.SriovNetSupport = aws.String("simple")
	}
	if s.EnableAMIENASupport {
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
	if err := awscommon.WaitUntilAMIAvailable(ec2conn, *registerResp.ImageId); err != nil {
		err := fmt.Errorf("Error waiting for AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("snapshots", map[string][]string{})

	return multistep.ActionContinue
}

func (s *StepRegisterAMI) Cleanup(multistep.StateBag) {}
