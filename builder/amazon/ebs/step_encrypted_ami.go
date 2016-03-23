package ebs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	awscommon "github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/packer"
)

type stepCreateEncryptedAMICopy struct {
	image *ec2.Image
}

func (s *stepCreateEncryptedAMICopy) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(Config)
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	// Encrypt boot not set, so skip step
	if !config.AMIConfig.AMIEncryptBootVolume {
		return multistep.ActionContinue
	}

	ui.Say("Creating Encrypted AMI Copy")

	amis := state.Get("amis").(map[string]string)
	var region, id string
	if amis != nil {
		for region, id = range amis {
			break // Only get the first
		}
	}

	ui.Say(fmt.Sprintf("Copying AMI: %s(%s)", region, id))

	copyOpts := &ec2.CopyImageInput{
		Name:          &config.AMIName, // Try to overwrite existing AMI
		SourceImageId: aws.String(id),
		SourceRegion:  aws.String(region),
		Encrypted:     aws.Bool(true),
	}

	copyResp, err := ec2conn.CopyImage(copyOpts)
	if err != nil {
		err := fmt.Errorf("Error copying AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Wait for the copy to become ready
	stateChange := awscommon.StateChangeConf{
		Pending:   []string{"pending"},
		Target:    "available",
		Refresh:   awscommon.AMIStateRefreshFunc(ec2conn, *copyResp.ImageId),
		StepState: state,
	}

	ui.Say("Waiting for AMI copy to become ready...")
	if _, err := awscommon.WaitForState(&stateChange); err != nil {
		err := fmt.Errorf("Error waiting for AMI Copy: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Remove unencrypted AMI
	ui.Say("Deregistering unecrypted AMI")
	deregisterOpts := &ec2.DeregisterImageInput{ImageId: aws.String(id)}
	if _, err := ec2conn.DeregisterImage(deregisterOpts); err != nil {
		ui.Error(fmt.Sprintf("Error deregistering AMI, may still be around: %s", err))
		return multistep.ActionHalt
	}

	// Replace original AMI ID with Encrypted ID in state
	amis[region] = *copyResp.ImageId
	state.Put("amis", amis)

	imagesResp, err := ec2conn.DescribeImages(&ec2.DescribeImagesInput{ImageIds: []*string{copyResp.ImageId}})
	if err != nil {
		err := fmt.Errorf("Error searching for AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	s.image = imagesResp.Images[0]

	return multistep.ActionContinue
}

func (s *stepCreateEncryptedAMICopy) Cleanup(state multistep.StateBag) {
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
	deregisterOpts := &ec2.DeregisterImageInput{ImageId: s.image.ImageId}
	if _, err := ec2conn.DeregisterImage(deregisterOpts); err != nil {
		ui.Error(fmt.Sprintf("Error deregistering AMI, may still be around: %s", err))
		return
	}
}
