package common

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepModifyAMIAttributes struct {
	Users        []string
	Groups       []string
	ProductCodes []string
	Description  string
}

func (s *StepModifyAMIAttributes) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)
	amis := state.Get("amis").(map[string]string)

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
	options := make(map[string]*ec2.ModifyImageAttributeInput)
	if s.Description != "" {
		options["description"] = &ec2.ModifyImageAttributeInput{
			Description: &ec2.AttributeValue{Value: &s.Description},
		}
	}

	if len(s.Groups) > 0 {
		groups := make([]*string, len(s.Groups))
		for i, g := range s.Groups {
			groups[i] = &g
		}
		options["groups"] = &ec2.ModifyImageAttributeInput{
			UserGroups: groups,
		}
	}

	if len(s.Users) > 0 {
		users := make([]*string, len(s.Users))
		for i, u := range s.Users {
			users[i] = &u
		}
		options["users"] = &ec2.ModifyImageAttributeInput{
			UserIDs: users,
		}
	}

	if len(s.ProductCodes) > 0 {
		codes := make([]*string, len(s.ProductCodes))
		for i, c := range s.ProductCodes {
			codes[i] = &c
		}
		options["product codes"] = &ec2.ModifyImageAttributeInput{
			ProductCodes: codes,
		}
	}

	for region, ami := range amis {
		ui.Say(fmt.Sprintf("Modifying attributes on AMI (%s)...", ami))
		regionconn := ec2.New(&aws.Config{
			Credentials: ec2conn.Config.Credentials,
			Region:      region,
		})
		for name, input := range options {
			ui.Message(fmt.Sprintf("Modifying: %s", name))
			input.ImageID = &ami
			_, err := regionconn.ModifyImageAttribute(input)
			if err != nil {
				err := fmt.Errorf("Error modify AMI attributes: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	return multistep.ActionContinue
}

func (s *StepModifyAMIAttributes) Cleanup(state multistep.StateBag) {
	// No cleanup...
}
