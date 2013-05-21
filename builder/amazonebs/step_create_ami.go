package amazonebs

import (
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/packer/packer"
)

type stepCreateAMI struct {}

func (s *stepCreateAMI) Run(state map[string]interface{}) StepAction {
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
		return StepHalt
	}

	ui.Say("AMI: %s", createResp.ImageId)

	// Wait for the image to become ready
	ui.Say("Waiting for AMI to become ready...")
	for {
		imageResp, err := ec2conn.Images([]string{createResp.ImageId}, ec2.NewFilter())
		if err != nil {
			ui.Error(err.Error())
			return StepHalt
		}

		if imageResp.Images[0].State == "available" {
			break
		}
	}

	return StepContinue
}

func (s *stepCreateAMI) Cleanup(map[string]interface{}) {
	// No cleanup...
}
