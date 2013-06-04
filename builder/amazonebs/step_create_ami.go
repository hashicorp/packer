package amazonebs

import (
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepCreateAMI struct{}

func (s *stepCreateAMI) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(config)
	ec2conn := state["ec2"].(*ec2.EC2)
	instance := state["instance"].(*ec2.Instance)
	ui := state["ui"].(packer.Ui)

	// Create the image
	ui.Say("Creating the AMI...")
	createOpts := &ec2.CreateImage{
		InstanceId: instance.InstanceId,
		Name:       config.AMIName,
	}

	createResp, err := ec2conn.CreateImage(createOpts)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the AMI ID in the state
	ui.Say(fmt.Sprintf("AMI: %s", createResp.ImageId))
	amis := make(map[string]string)
	amis[config.Region] = createResp.ImageId
	state["amis"] = amis

	// Wait for the image to become ready
	ui.Say("Waiting for AMI to become ready...")
	for {
		imageResp, err := ec2conn.Images([]string{createResp.ImageId}, ec2.NewFilter())
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if imageResp.Images[0].State == "available" {
			break
		}
	}

	return multistep.ActionContinue
}

func (s *stepCreateAMI) Cleanup(map[string]interface{}) {
	// No cleanup...
}
