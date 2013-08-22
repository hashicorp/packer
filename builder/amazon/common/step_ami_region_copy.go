package common

import (
	"fmt"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepAMIRegionCopy struct {
	Regions []string
	Tags    map[string]string
}

func (s *StepAMIRegionCopy) Run(state map[string]interface{}) multistep.StepAction {
	ec2conn := state["ec2"].(*ec2.EC2)
	ui := state["ui"].(packer.Ui)
	amis := state["amis"].(map[string]string)
	ami := amis[ec2conn.Region.Name]

	if len(s.Regions) == 0 {
		return multistep.ActionContinue
	}

	for _, region := range s.Regions {
		ui.Say(fmt.Sprintf("Copying AMI (%s) to region (%s)...", ami, region))

		// Connect to the region where the AMI will be copied to
		regionconn := ec2.New(ec2conn.Auth, aws.Regions[region])
		resp, err := regionconn.CopyImage(&ec2.CopyImage{
			SourceRegion:  ec2conn.Region.Name,
			SourceImageId: ami,
		})

		if err != nil {
			err := fmt.Errorf("Error Copying AMI (%s) to region (%s): %s", ami, region, err)
			state["error"] = err
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		ui.Say(fmt.Sprintf("Waiting for AMI (%s) in region (%s) to become ready...", resp.ImageId, region))
		if err := WaitForAMI(regionconn, resp.ImageId); err != nil {
			err := fmt.Errorf("Error waiting for AMI (%s) in region (%s): %s", resp.ImageId, region, err)
			state["error"] = err
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Need to re-apply Tags since they are not copied with the AMI
		if len(s.Tags) > 0 {
			ui.Say(fmt.Sprintf("Adding tags to AMI (%s)...", resp.ImageId))

			var ec2Tags []ec2.Tag
			for key, value := range s.Tags {
				ui.Message(fmt.Sprintf("Adding tag: \"%s\": \"%s\"", key, value))
				ec2Tags = append(ec2Tags, ec2.Tag{key, value})
			}

			_, err := regionconn.CreateTags([]string{resp.ImageId}, ec2Tags)
			if err != nil {
				err := fmt.Errorf("Error adding tags to AMI (%s): %s", resp.ImageId, err)
				state["error"] = err
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}

		amis[region] = resp.ImageId
	}

	state["amis"] = amis
	return multistep.ActionContinue
}

func (s *StepAMIRegionCopy) Cleanup(state map[string]interface{}) {
	// No cleanup...
}
