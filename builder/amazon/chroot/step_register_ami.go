package chroot

import (
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	awscommon "github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/packer"
)

// StepRegisterAMI creates the AMI.
type StepRegisterAMI struct{}

func (s *StepRegisterAMI) Run(state map[string]interface{}) multistep.StepAction {
	ec2conn := state["ec2"].(*ec2.EC2)
	image := state["source_image"].(*ec2.Image)
	snapshotId := state["snapshot_id"].(string)
	ui := state["ui"].(packer.Ui)

	amiName := "foo"

	ui.Say("Registering the AMI...")
	blockDevices := make([]ec2.BlockDeviceMapping, len(image.BlockDevices))
	for i, device := range image.BlockDevices {
		newDevice := device
		if newDevice.DeviceName == image.RootDeviceName {
			newDevice.SnapshotId = snapshotId
		}

		blockDevices[i] = newDevice
	}

	registerOpts := &ec2.RegisterImage{
		Name:           amiName,
		Architecture:   image.Architecture,
		KernelId:       image.KernelId,
		RamdiskId:      image.RamdiskId,
		RootDeviceName: image.RootDeviceName,
		BlockDevices:   blockDevices,
	}

	registerResp, err := ec2conn.RegisterImage(registerOpts)
	if err != nil {
		state["error"] = fmt.Errorf("Error registering AMI: %s", err)
		ui.Error(state["error"].(error).Error())
		return multistep.ActionHalt
	}

	// Set the AMI ID in the state
	ui.Say(fmt.Sprintf("AMI: %s", registerResp.ImageId))
	amis := make(map[string]string)
	amis[ec2conn.Region.Name] = registerResp.ImageId
	state["amis"] = amis

	// Wait for the image to become ready
	ui.Say("Waiting for AMI to become ready...")
	if err := awscommon.WaitForAMI(ec2conn, registerResp.ImageId); err != nil {
		err := fmt.Errorf("Error waiting for AMI: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepRegisterAMI) Cleanup(state map[string]interface{}) {}
