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

	options := &ec2.ModifyImageAttribute{
		Description:  s.Description,
		AddUsers:     s.Users,
		AddGroups:    s.Groups,
		ProductCodes: s.ProductCodes,
	}

	ui.Say("Modifying AMI attributes...")
	_, err := ec2conn.ModifyImageAttribute(ami, options)
	if err != nil {
		err := fmt.Errorf("Error modify AMI attributes: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepModifyAttributes) Cleanup(state map[string]interface{}) {
	// No cleanup...
}
