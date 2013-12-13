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

func (s *StepRegisterAMI) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ec2conn := state.Get("ec2").(*ec2.EC2)
	image := state.Get("source_image").(*ec2.Image)
	snapshotId := state.Get("snapshot_id").(string)
	ui := state.Get("ui").(packer.Ui)

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
		Name:           config.AMIName,
		Architecture:   image.Architecture,
		KernelId:       image.KernelId,
		RamdiskId:      image.RamdiskId,
		RootDeviceName: image.RootDeviceName,
		BlockDevices:   blockDevices,
	}

	registerResp, err := ec2conn.RegisterImage(registerOpts)
	if err != nil {
		state.Put("error", fmt.Errorf("Error registering AMI: %s", err))
		ui.Error(state.Get("error").(error).Error())
		return multistep.ActionHalt
	}

	// Set the AMI ID in the state
	ui.Say(fmt.Sprintf("AMI: %s", registerResp.ImageId))
	amis := make(map[string]string)
	amis[ec2conn.Region.Name] = registerResp.ImageId
	state.Put("amis", amis)

	// Wait for the image to become ready
	stateChange := awscommon.StateChangeConf{
		Pending:   []string{"pending"},
		Target:    "available",
		Refresh:   awscommon.AMIStateRefreshFunc(ec2conn, registerResp.ImageId),
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
