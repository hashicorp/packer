package chroot

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// StepCheckRootDevice makes sure the root device on the AMI is EBS-backed.
type StepCheckRootDevice struct{}

func (s *StepCheckRootDevice) Run(state multistep.StateBag) multistep.StepAction {
	image := state.Get("source_image").(*ec2.Image)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Checking the root device on source AMI...")

	// It must be EBS-backed otherwise the build won't work
	if *image.RootDeviceType != "ebs" {
		err := fmt.Errorf("The root device of the source AMI must be EBS-backed.")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepCheckRootDevice) Cleanup(multistep.StateBag) {}
