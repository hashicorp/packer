package common

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepDeregisterAMI struct {
	ForceDeregister bool
	AMIName         string
}

func (s *StepDeregisterAMI) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	// check for force deregister
	if s.ForceDeregister {
		resp, err := ec2conn.DescribeImages(&ec2.DescribeImagesInput{
			Filters: []*ec2.Filter{&ec2.Filter{
				Name:   aws.String("name"),
				Values: []*string{aws.String(s.AMIName)},
			}}})

		if err != nil {
			err := fmt.Errorf("Error creating AMI: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// deregister image(s) by that name
		for _, i := range resp.Images {
			_, err := ec2conn.DeregisterImage(&ec2.DeregisterImageInput{
				ImageId: i.ImageId,
			})

			if err != nil {
				err := fmt.Errorf("Error deregistering existing AMI: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
			ui.Say(fmt.Sprintf("Deregistered AMI %s, id: %s", s.AMIName, *i.ImageId))
		}
	}

	return multistep.ActionContinue
}

func (s *StepDeregisterAMI) Cleanup(state multistep.StateBag) {
}
