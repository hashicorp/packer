package common

import (
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepModifyAMIAttributes struct {
	Users        []string
	Groups       []string
	ProductCodes []string
	Description  string
}

func (s *StepModifyAMIAttributes) Run(state map[string]interface{}) multistep.StepAction {
	ec2conn := state["ec2"].(*ec2.EC2)
	ui := state["ui"].(packer.Ui)
	amis := state["amis"].(map[string]string)
	ami := amis[ec2conn.Region.Name]

	// Determine if there is any work to do.
	valid := false
	valid = valid || s.Description != ""
	valid = valid || (s.Users != nil && len(s.Users) > 0)
	valid = valid || (s.Groups != nil && len(s.Groups) > 0)
	valid = valid || (s.ProductCodes != nil && len(s.ProductCodes) > 0)

	if !valid {
		return multistep.ActionContinue
	}

	// Construct the modify image attribute requests we're going to make.
	// We need to make each separately since the EC2 API only allows changing
	// one type at a kind currently.
	options := make(map[string]*ec2.ModifyImageAttribute)
	if s.Description != "" {
		options["description"] = &ec2.ModifyImageAttribute{
			Description: s.Description,
		}
	}

	if len(s.Groups) > 0 {
		options["groups"] = &ec2.ModifyImageAttribute{
			AddGroups: s.Groups,
		}
	}

	if len(s.Users) > 0 {
		options["users"] = &ec2.ModifyImageAttribute{
			AddUsers: s.Users,
		}
	}

	if len(s.ProductCodes) > 0 {
		options["product codes"] = &ec2.ModifyImageAttribute{
			ProductCodes: s.ProductCodes,
		}
	}

	ui.Say("Modifying AMI attributes...")
	for name, opts := range options {
		ui.Message(fmt.Sprintf("Modifying: %s", name))
		_, err := ec2conn.ModifyImageAttribute(ami, opts)
		if err != nil {
			err := fmt.Errorf("Error modify AMI attributes: %s", err)
			state["error"] = err
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepModifyAMIAttributes) Cleanup(state map[string]interface{}) {
	// No cleanup...
}
