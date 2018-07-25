package common

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepPreValidate provides an opportunity to pre-validate any configuration for
// the build before actually doing any time consuming work
//
type StepPreValidate struct {
	DestAmiName     string
	ForceDeregister bool
}

func (s *StepPreValidate) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	if s.ForceDeregister {
		ui.Say("Force Deregister flag found, skipping prevalidating AMI Name")
		return multistep.ActionContinue
	}

	ec2conn := state.Get("ec2").(*ec2.EC2)

	ui.Say(fmt.Sprintf("Prevalidating AMI Name: %s", s.DestAmiName))
	resp, err := ec2conn.DescribeImages(&ec2.DescribeImagesInput{
		Filters: []*ec2.Filter{{
			Name:   aws.String("name"),
			Values: []*string{aws.String(s.DestAmiName)},
		}}})

	if err != nil {
		err := fmt.Errorf("Error querying AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(resp.Images) > 0 {
		err := fmt.Errorf("Error: name conflicts with an existing AMI: %s", *resp.Images[0].ImageId)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepPreValidate) Cleanup(multistep.StateBag) {}
