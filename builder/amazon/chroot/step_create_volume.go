package chroot

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepCreateVolume creates a new volume from the snapshot of the root
// device of the AMI.
//
// Produces:
//   volume_id string - The ID of the created volume
type StepCreateVolume struct {
	volumeId       string
	RootVolumeSize int64
}

func (s *StepCreateVolume) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ec2conn := state.Get("ec2").(*ec2.EC2)
	instance := state.Get("instance").(*ec2.Instance)
	ui := state.Get("ui").(packer.Ui)

	var createVolume *ec2.CreateVolumeInput
	if config.FromScratch {
		createVolume = &ec2.CreateVolumeInput{
			AvailabilityZone: instance.Placement.AvailabilityZone,
			Size:             aws.Int64(s.RootVolumeSize),
			VolumeType:       aws.String(ec2.VolumeTypeGp2),
		}
	} else {
		// Determine the root device snapshot
		image := state.Get("source_image").(*ec2.Image)
		log.Printf("Searching for root device of the image (%s)", *image.RootDeviceName)
		var rootDevice *ec2.BlockDeviceMapping
		for _, device := range image.BlockDeviceMappings {
			if *device.DeviceName == *image.RootDeviceName {
				rootDevice = device
				break
			}
		}

		if rootDevice == nil {
			err := fmt.Errorf("Couldn't find root device!")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		ui.Say("Creating the root volume...")
		vs := *rootDevice.Ebs.VolumeSize
		if s.RootVolumeSize > *rootDevice.Ebs.VolumeSize {
			vs = s.RootVolumeSize
		}

		createVolume = &ec2.CreateVolumeInput{
			AvailabilityZone: instance.Placement.AvailabilityZone,
			Size:             aws.Int64(vs),
			SnapshotId:       rootDevice.Ebs.SnapshotId,
			VolumeType:       rootDevice.Ebs.VolumeType,
			Iops:             rootDevice.Ebs.Iops,
		}
	}

	log.Printf("Create args: %+v", createVolume)

	createVolumeResp, err := ec2conn.CreateVolume(createVolume)
	if err != nil {
		err := fmt.Errorf("Error creating root volume: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the volume ID so we remember to delete it later
	s.volumeId = *createVolumeResp.VolumeId
	log.Printf("Volume ID: %s", s.volumeId)

	// Wait for the volume to become ready
	err = awscommon.WaitUntilVolumeAvailable(ctx, ec2conn, s.volumeId)
	if err != nil {
		err := fmt.Errorf("Error waiting for volume: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("volume_id", s.volumeId)
	return multistep.ActionContinue
}

func (s *StepCreateVolume) Cleanup(state multistep.StateBag) {
	if s.volumeId == "" {
		return
	}

	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting the created EBS volume...")
	_, err := ec2conn.DeleteVolume(&ec2.DeleteVolumeInput{VolumeId: &s.volumeId})
	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting EBS volume: %s", err))
	}
}
