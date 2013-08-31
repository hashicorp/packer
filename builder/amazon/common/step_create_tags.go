package common

import (
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepCreateTags struct {
	Tags map[string]string
}

func (s *StepCreateTags) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)
	amis := state.Get("amis").(map[string]string)
	ami := amis[ec2conn.Region.Name]

	if len(s.Tags) > 0 {
		ui.Say(fmt.Sprintf("Adding tags to AMI (%s)...", ami))

		var ec2Tags []ec2.Tag
		for key, value := range s.Tags {
			ui.Message(fmt.Sprintf("Adding tag: \"%s\": \"%s\"", key, value))
			ec2Tags = append(ec2Tags, ec2.Tag{key, value})
		}

		_, err := ec2conn.CreateTags([]string{ami}, ec2Tags)
		if err != nil {
			err := fmt.Errorf("Error adding tags to AMI (%s): %s", ami, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepCreateTags) Cleanup(state multistep.StateBag) {
	// No cleanup...
}
