package chroot

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

// StepCreateVolume creates a new volume from the snapshot of the root
// device of the AMI.
//
// Produces:
//   volume_id string - The ID of the created volume
type StepCreateVolume struct {
	volumeId       string
	RootVolumeSize int64
	RootVolumeType string
	RootVolumeTags awscommon.TagMap
	Ctx            interpolate.Context
}

func (s *StepCreateVolume) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ec2conn := state.Get("ec2").(*ec2.EC2)
	instance := state.Get("instance").(*ec2.Instance)
	ui := state.Get("ui").(packer.Ui)

	volTags, err := s.RootVolumeTags.EC2Tags(s.Ctx, *ec2conn.Config.Region, state)
	if err != nil {
		err := fmt.Errorf("Error tagging volumes: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Collect tags for tagging on resource creation
	var tagSpecs []*ec2.TagSpecification

	if len(volTags) > 0 {
		runVolTags := &ec2.TagSpecification{
			ResourceType: aws.String("volume"),
			Tags:         volTags,
		}

		tagSpecs = append(tagSpecs, runVolTags)
	}

	var createVolume *ec2.CreateVolumeInput
	if config.FromScratch {
		rootVolumeType := ec2.VolumeTypeGp2
		if s.RootVolumeType == "io1" {
			err := errors.New("Cannot use io1 volume when building from scratch")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		} else if s.RootVolumeType != "" {
			rootVolumeType = s.RootVolumeType
		}
		createVolume = &ec2.CreateVolumeInput{
			AvailabilityZone: instance.Placement.AvailabilityZone,
			Size:             aws.Int64(s.RootVolumeSize),
			VolumeType:       aws.String(rootVolumeType),
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

		ui.Say("Creating the root volume...")
		createVolume, err = s.buildCreateVolumeInput(*instance.Placement.AvailabilityZone, rootDevice)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	if len(tagSpecs) > 0 {
		createVolume.SetTagSpecifications(tagSpecs)
		volTags.Report(ui)
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

func (s *StepCreateVolume) buildCreateVolumeInput(az string, rootDevice *ec2.BlockDeviceMapping) (*ec2.CreateVolumeInput, error) {
	if rootDevice == nil {
		return nil, fmt.Errorf("Couldn't find root device!")
	}
	createVolumeInput := &ec2.CreateVolumeInput{
		AvailabilityZone: aws.String(az),
		Size:             rootDevice.Ebs.VolumeSize,
		SnapshotId:       rootDevice.Ebs.SnapshotId,
		VolumeType:       rootDevice.Ebs.VolumeType,
		Iops:             rootDevice.Ebs.Iops,
	}
	if s.RootVolumeSize > *rootDevice.Ebs.VolumeSize {
		createVolumeInput.Size = aws.Int64(s.RootVolumeSize)
	}

	if s.RootVolumeType == "" || s.RootVolumeType == *rootDevice.Ebs.VolumeType {
		return createVolumeInput, nil
	}

	if s.RootVolumeType == "io1" {
		return nil, fmt.Errorf("Root volume type cannot be io1, because existing root volume type was %s", *rootDevice.Ebs.VolumeType)
	}

	createVolumeInput.VolumeType = aws.String(s.RootVolumeType)
	// non io1 cannot set iops
	createVolumeInput.Iops = nil

	return createVolumeInput, nil
}
