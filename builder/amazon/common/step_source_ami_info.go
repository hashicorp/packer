package common

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// StepSourceAMIInfo extracts critical information from the source AMI
// that is used throughout the AMI creation process.
//
// Produces:
//   source_image *ec2.Image - the source AMI info
type StepSourceAMIInfo struct {
	SourceAmi          string
	EnhancedNetworking bool
}

func (s *StepSourceAMIInfo) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	ui.Say(fmt.Sprintf("Inspecting the source AMI (%s)...", s.SourceAmi))
	imageResp, err := ec2conn.DescribeImages(&ec2.DescribeImagesInput{ImageIds: []*string{&s.SourceAmi}})
	if err != nil {
		err := fmt.Errorf("Error querying AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(imageResp.Images) == 0 {
		err := fmt.Errorf("Source AMI '%s' was not found!", s.SourceAmi)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	image := imageResp.Images[0]

	// Enhanced Networking (SriovNetSupport) can only be enabled on HVM AMIs.
	// See http://goo.gl/icuXh5
	if s.EnhancedNetworking && *image.VirtualizationType != "hvm" {
		err := fmt.Errorf("Cannot enable enhanced networking, source AMI '%s' is not HVM", s.SourceAmi)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("source_image", image)
	return multistep.ActionContinue
}

func (s *StepSourceAMIInfo) Cleanup(multistep.StateBag) {}
