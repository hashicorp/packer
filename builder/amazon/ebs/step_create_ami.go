package ebs

import (
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	awscommon "github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/packer"
)

type stepCreateAMI struct{}

func (s *stepCreateAMI) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(config)
	ec2conn := state["ec2"].(*ec2.EC2)
	instance := state["instance"].(*ec2.Instance)
	ui := state["ui"].(packer.Ui)

	// Create the image
	ui.Say(fmt.Sprintf("Creating the AMI: %s", config.AMIName))
	createOpts := &ec2.CreateImage{
		InstanceId:   instance.InstanceId,
		Name:         config.AMIName,
		BlockDevices: config.BlockDevices.BuildAMIDevices(),
	}

	createResp, err := ec2conn.CreateImage(createOpts)
	if err != nil {
		err := fmt.Errorf("Error creating AMI: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the AMI ID in the state
	ui.Say(fmt.Sprintf("AMI: %s", createResp.ImageId))
	amis := make(map[string]string)
	amis[ec2conn.Region.Name] = createResp.ImageId
	state["amis"] = amis

	// Wait for the image to become ready
	ui.Say("Waiting for AMI to become ready...")
	if err := awscommon.WaitForAMI(ec2conn, createResp.ImageId); err != nil {
		err := fmt.Errorf("Error waiting for AMI: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepCreateAMI) Cleanup(map[string]interface{}) {
	// No cleanup...
}
