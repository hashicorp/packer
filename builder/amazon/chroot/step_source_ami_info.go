package chroot

import (
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// StepSourceAMIInfo extracts critical information from the source AMI
// that is used throughout the AMI creation process.
//
// Produces:
//   source_image *ec2.Image - the source AMI info
type StepSourceAMIInfo struct{}

func (s *StepSourceAMIInfo) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Inspecting the source AMI...")
	imageResp, err := ec2conn.Images([]string{config.SourceAmi}, ec2.NewFilter())
	if err != nil {
		err := fmt.Errorf("Error querying AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(imageResp.Images) == 0 {
		err := fmt.Errorf("Source AMI '%s' was not found!", config.SourceAmi)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	image := &imageResp.Images[0]

	// It must be EBS-backed otherwise the build won't work
	if image.RootDeviceType != "ebs" {
		err := fmt.Errorf("The root device of the source AMI must be EBS-backed.")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("source_image", image)
	return multistep.ActionContinue
}

func (s *StepSourceAMIInfo) Cleanup(multistep.StateBag) {}
