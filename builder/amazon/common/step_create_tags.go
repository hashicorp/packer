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

func (s *StepCreateTags) Run(state map[string]interface{}) multistep.StepAction {
	ec2conn := state["ec2"].(*ec2.EC2)
	ui := state["ui"].(packer.Ui)
	amis := state["amis"].(map[string]string)
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
			state["error"] = err
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepCreateTags) Cleanup(state map[string]interface{}) {
	// No cleanup...
}
