package ebs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	awscommon "github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/packer"
)

type stepCreateAMI struct {
	image *ec2.Image
}

func (s *stepCreateAMI) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(Config)
	ec2conn := state.Get("ec2").(*ec2.EC2)
	instance := state.Get("instance").(*ec2.Instance)
	ui := state.Get("ui").(packer.Ui)

	// Create the image
	ui.Say(fmt.Sprintf("Creating the AMI: %s", config.AMIName))
	createOpts := &ec2.CreateImageInput{
		InstanceID:          instance.InstanceID,
		Name:                &config.AMIName,
		BlockDeviceMappings: config.BlockDevices.BuildAMIDevices(),
	}

	createResp, err := ec2conn.CreateImage(createOpts)
	if err != nil {
		err := fmt.Errorf("Error creating AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the AMI ID in the state
	ui.Message(fmt.Sprintf("AMI: %s", *createResp.ImageID))
	amis := make(map[string]string)
	amis[ec2conn.Config.Region] = *createResp.ImageID
	state.Put("amis", amis)

	// Wait for the image to become ready
	stateChange := awscommon.StateChangeConf{
		Pending:   []string{"pending"},
		Target:    "available",
		Refresh:   awscommon.AMIStateRefreshFunc(ec2conn, *createResp.ImageID),
		StepState: state,
	}

	ui.Say("Waiting for AMI to become ready...")
	if _, err := awscommon.WaitForState(&stateChange); err != nil {
		err := fmt.Errorf("Error waiting for AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	imagesResp, err := ec2conn.DescribeImages(&ec2.DescribeImagesInput{ImageIDs: []*string{createResp.ImageID}})
	if err != nil {
		err := fmt.Errorf("Error searching for AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	s.image = imagesResp.Images[0]

	return multistep.ActionContinue
}

func (s *stepCreateAMI) Cleanup(state multistep.StateBag) {
	if s.image == nil {
		return
	}

	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deregistering the AMI because cancelation or error...")
	deregisterOpts := &ec2.DeregisterImageInput{ImageID: s.image.ImageID}
	if _, err := ec2conn.DeregisterImage(deregisterOpts); err != nil {
		ui.Error(fmt.Sprintf("Error deregistering AMI, may still be around: %s", err))
		return
	}
}
