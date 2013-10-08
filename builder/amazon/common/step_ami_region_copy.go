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
}

func (s *StepAMIRegionCopy) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)
	amis := state.Get("amis").(map[string]string)
	ami := amis[ec2conn.Region.Name]

	if len(s.Regions) == 0 {
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Copying AMI (%s) to other regions...", ami))
	for _, region := range s.Regions {
		ui.Message(fmt.Sprintf("Copying to: %s", region))

		// Connect to the region where the AMI will be copied to
		regionconn := ec2.New(ec2conn.Auth, aws.Regions[region])
		resp, err := regionconn.CopyImage(&ec2.CopyImage{
			SourceRegion:  ec2conn.Region.Name,
			SourceImageId: ami,
		})

		if err != nil {
			err := fmt.Errorf("Error Copying AMI (%s) to region (%s): %s", ami, region, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		stateChange := StateChangeConf{
			Conn:      regionconn,
			Pending:   []string{"pending"},
			Target:    "available",
			Refresh:   AMIStateRefreshFunc(regionconn, resp.ImageId),
			StepState: state,
		}

		ui.Say(fmt.Sprintf("Waiting for AMI (%s) in region (%s) to become ready...",
			resp.ImageId, region))
		if _, err := WaitForState(&stateChange); err != nil {
			err := fmt.Errorf("Error waiting for AMI (%s) in region (%s): %s", resp.ImageId, region, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		amis[region] = resp.ImageId
	}

	state.Put("amis", amis)
	return multistep.ActionContinue
}

func (s *StepAMIRegionCopy) Cleanup(state multistep.StateBag) {
	// No cleanup...
}
