package chroot

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	awscommon "github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/packer"
)

// StepRegisterAMI creates the AMI.
type StepRegisterAMI struct {
	RootVolumeSize int64
}

func (s *StepRegisterAMI) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ec2conn := state.Get("ec2").(*ec2.EC2)
	image := state.Get("source_image").(*ec2.Image)
	snapshotId := state.Get("snapshot_id").(string)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Registering the AMI...")
	blockDevices := make([]*ec2.BlockDeviceMapping, len(image.BlockDeviceMappings))
	for i, device := range image.BlockDeviceMappings {
		newDevice := device
		if *newDevice.DeviceName == *image.RootDeviceName {
			if newDevice.Ebs != nil {
				newDevice.Ebs.SnapshotId = aws.String(snapshotId)
			} else {
				newDevice.Ebs = &ec2.EbsBlockDevice{SnapshotId: aws.String(snapshotId)}
			}

			if s.RootVolumeSize > *newDevice.Ebs.VolumeSize {
				newDevice.Ebs.VolumeSize = aws.Int64(s.RootVolumeSize)
			}
		}

		// assume working from a snapshot, so we unset the Encrypted field if set,
		// otherwise AWS API will return InvalidParameter
		if newDevice.Ebs != nil && newDevice.Ebs.Encrypted != nil {
			newDevice.Ebs.Encrypted = nil
		}

		blockDevices[i] = newDevice
	}

	registerOpts := buildRegisterOpts(config, image, blockDevices)

	// Set SriovNetSupport to "simple". See http://goo.gl/icuXh5
	if config.AMIEnhancedNetworking {
		registerOpts.SriovNetSupport = aws.String("simple")
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
	stateChange := awscommon.StateChangeConf{
		Pending:   []string{"pending"},
		Target:    "available",
		Refresh:   awscommon.AMIStateRefreshFunc(ec2conn, *registerResp.ImageId),
		StepState: state,
	}

	ui.Say("Waiting for AMI to become ready...")
	if _, err := awscommon.WaitForState(&stateChange); err != nil {
		err := fmt.Errorf("Error waiting for AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepRegisterAMI) Cleanup(state multistep.StateBag) {}

func buildRegisterOpts(config *Config, image *ec2.Image, blockDevices []*ec2.BlockDeviceMapping) *ec2.RegisterImageInput {
	registerOpts := &ec2.RegisterImageInput{
		Name:                &config.AMIName,
		Architecture:        image.Architecture,
		RootDeviceName:      image.RootDeviceName,
		BlockDeviceMappings: blockDevices,
		VirtualizationType:  image.VirtualizationType,
	}

	if config.AMIVirtType != "" {
		registerOpts.VirtualizationType = aws.String(config.AMIVirtType)
	}

	if config.AMIVirtType != "hvm" {
		registerOpts.KernelId = image.KernelId
		registerOpts.RamdiskId = image.RamdiskId
	}

	return registerOpts
}
