package chroot

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepRegisterAMI creates the AMI.
type StepRegisterAMI struct {
	RootVolumeSize           int64
	EnableAMIENASupport      bool
	EnableAMISriovNetSupport bool
}

func (s *StepRegisterAMI) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ec2conn := state.Get("ec2").(*ec2.EC2)
	snapshotId := state.Get("snapshot_id").(string)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Registering the AMI...")

	var (
		registerOpts   *ec2.RegisterImageInput
		mappings       []*ec2.BlockDeviceMapping
		image          *ec2.Image
		rootDeviceName string
	)

	if config.FromScratch {
		mappings = config.AMIBlockDevices.BuildAMIDevices()
		rootDeviceName = config.RootDeviceName
	} else {
		image = state.Get("source_image").(*ec2.Image)
		mappings = image.BlockDeviceMappings
		rootDeviceName = *image.RootDeviceName
	}

	newMappings := make([]*ec2.BlockDeviceMapping, len(mappings))
	for i, device := range mappings {
		newDevice := device
		if *newDevice.DeviceName == rootDeviceName {
			if newDevice.Ebs != nil {
				newDevice.Ebs.SnapshotId = aws.String(snapshotId)
			} else {
				newDevice.Ebs = &ec2.EbsBlockDevice{SnapshotId: aws.String(snapshotId)}
			}

			if config.FromScratch || s.RootVolumeSize > *newDevice.Ebs.VolumeSize {
				newDevice.Ebs.VolumeSize = aws.Int64(s.RootVolumeSize)
			}
		}

		// assume working from a snapshot, so we unset the Encrypted field if set,
		// otherwise AWS API will return InvalidParameter
		if newDevice.Ebs != nil && newDevice.Ebs.Encrypted != nil {
			newDevice.Ebs.Encrypted = nil
		}

		newMappings[i] = newDevice
	}

	if config.FromScratch {
		registerOpts = &ec2.RegisterImageInput{
			Name:                &config.AMIName,
			Architecture:        aws.String(ec2.ArchitectureValuesX8664),
			RootDeviceName:      aws.String(rootDeviceName),
			VirtualizationType:  aws.String(config.AMIVirtType),
			BlockDeviceMappings: newMappings,
		}
	} else {
		registerOpts = buildRegisterOpts(config, image, newMappings)
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
	if err := awscommon.WaitUntilAMIAvailable(ec2conn, *registerResp.ImageId); err != nil {
		err := fmt.Errorf("Error waiting for AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Waiting for AMI to become ready...")

	return multistep.ActionContinue
}

func (s *StepRegisterAMI) Cleanup(state multistep.StateBag) {}

func buildRegisterOpts(config *Config, image *ec2.Image, mappings []*ec2.BlockDeviceMapping) *ec2.RegisterImageInput {
	registerOpts := &ec2.RegisterImageInput{
		Name:                &config.AMIName,
		Architecture:        image.Architecture,
		RootDeviceName:      image.RootDeviceName,
		BlockDeviceMappings: mappings,
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
