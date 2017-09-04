package common

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type StepDeregisterAMI struct {
	AccessConfig        *AccessConfig
	ForceDeregister     bool
	ForceDeleteSnapshot bool
	AMIName             string
	Regions             []string
}

func (s *StepDeregisterAMI) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	regions := s.Regions
	if len(regions) == 0 {
		regions = append(regions, s.AccessConfig.RawRegion)
	}

	// Check for force deregister
	if s.ForceDeregister {
		for _, region := range regions {
			// get new connection for each region in which we need to deregister vms
			session, err := s.AccessConfig.Session()
			if err != nil {
				return multistep.ActionHalt
			}

			regionconn := ec2.New(session.Copy(&aws.Config{
				Region: aws.String(region)},
			))

			resp, err := regionconn.DescribeImages(&ec2.DescribeImagesInput{
				Filters: []*ec2.Filter{{
					Name:   aws.String("name"),
					Values: []*string{aws.String(s.AMIName)},
				}}})

			if err != nil {
				err := fmt.Errorf("Error describing AMI: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}

			// Deregister image(s) by name
			for _, i := range resp.Images {
				_, err := regionconn.DeregisterImage(&ec2.DeregisterImageInput{
					ImageId: i.ImageId,
				})

				if err != nil {
					err := fmt.Errorf("Error deregistering existing AMI: %s", err)
					state.Put("error", err)
					ui.Error(err.Error())
					return multistep.ActionHalt
				}
				ui.Say(fmt.Sprintf("Deregistered AMI %s, id: %s", s.AMIName, *i.ImageId))

				// Delete snapshot(s) by image
				if s.ForceDeleteSnapshot {
					for _, b := range i.BlockDeviceMappings {
						if b.Ebs != nil && aws.StringValue(b.Ebs.SnapshotId) != "" {
							_, err := regionconn.DeleteSnapshot(&ec2.DeleteSnapshotInput{
								SnapshotId: b.Ebs.SnapshotId,
							})

							if err != nil {
								err := fmt.Errorf("Error deleting existing snapshot: %s", err)
								state.Put("error", err)
								ui.Error(err.Error())
								return multistep.ActionHalt
							}
							ui.Say(fmt.Sprintf("Deleted snapshot: %s", *b.Ebs.SnapshotId))
						}
					}
				}
			}
		}
	}

	return multistep.ActionContinue
}

func (s *StepDeregisterAMI) Cleanup(state multistep.StateBag) {
}
