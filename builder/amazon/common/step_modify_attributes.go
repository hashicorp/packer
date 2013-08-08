package common

import (
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepModifyAttributes struct {
	Users        []string
	Groups       []string
	ProductCodes []string
	Description  string
}

func (s *StepModifyAttributes) Run(state map[string]interface{}) multistep.StepAction {
	ec2conn := state["ec2"].(*ec2.EC2)
	ui := state["ui"].(packer.Ui)
	amis := state["amis"].(map[string]string)
	ami := amis[ec2conn.Region.Name]

	if s.Description != "" {
		ui.Say(fmt.Sprintf("Setting Description of AMI (%s) to '%s'...", ami, s.Description))
		_, err := ec2conn.ModifyImageAttribute(ami, &ec2.ModifyImageAttribute{
			Attribute:   ec2.DescriptionAttribute,
			Description: s.Description,
		})
		if err != nil {
			err := fmt.Errorf("Error setting Description of AMI (%s): %s", ami, err)
			state["error"] = err
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	if len(s.Users) > 0 || len(s.Groups) > 0 {
		ui.Say(fmt.Sprintf("Setting Launch Permissions for AMI (%s)...", ami))
		_, err := ec2conn.ModifyImageAttribute(ami, &ec2.ModifyImageAttribute{
			Attribute: ec2.LaunchPermissionAttribute,
			Operation: ec2.LaunchPermissionAdd,
			Users:     s.Users,
			Groups:    s.Groups,
		})
		if err != nil {
			err := fmt.Errorf("Error setting Launch Permissions for AMI (%s): %s", ami, err)
			state["error"] = err
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	if len(s.ProductCodes) > 0 {
		ui.Say(fmt.Sprintf("Setting Product Code(s) for AMI (%s)...", ami))
		_, err := ec2conn.ModifyImageAttribute(ami, &ec2.ModifyImageAttribute{
			Attribute:    ec2.ProductCodeAttribute,
			ProductCodes: s.ProductCodes,
		})
		if err != nil {
			err := fmt.Errorf("Error setting Product Code(s) for AMI (%s): %s", ami, err)
			state["error"] = err
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepModifyAttributes) Cleanup(state map[string]interface{}) {
	// No cleanup...
}
