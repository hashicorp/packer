package chroot

import (
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

// StepCreateVolume creates a new volume from the snapshot of the root
// device of the AMI.
//
// Produces:
//   volume_id string - The ID of the created volume
type StepCreateVolume struct {
	volumeId string
}

func (s *StepCreateVolume) Run(state map[string]interface{}) multistep.StepAction {
	ec2conn := state["ec2"].(*ec2.EC2)
	image := state["source_image"].(*ec2.Image)
	instance := state["instance"].(*ec2.Instance)
	ui := state["ui"].(packer.Ui)

	// Determine the root device snapshot
	log.Printf("Searching for root device of the image (%s)", image.RootDeviceName)
	var rootDevice *ec2.BlockDeviceMapping
	for _, device := range image.BlockDevices {
		if device.DeviceName == image.RootDeviceName {
			rootDevice = &device
			break
		}
	}

	if rootDevice == nil {
		err := fmt.Errorf("Couldn't find root device!")
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Creating the root volume...")
	createVolume := &ec2.CreateVolume{
		AvailZone:  instance.AvailZone,
		Size:       rootDevice.VolumeSize,
		SnapshotId: rootDevice.SnapshotId,
		VolumeType: rootDevice.VolumeType,
		IOPS:       rootDevice.IOPS,
	}

	createVolumeResp, err := ec2conn.CreateVolume(createVolume)
	if err != nil {
		err := fmt.Errorf("Error creating root volume: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the volume ID so we remember to delete it later
	s.volumeId = createVolumeResp.VolumeId

	return multistep.ActionContinue
}

func (s *StepCreateVolume) Cleanup(map[string]interface{}) {}
