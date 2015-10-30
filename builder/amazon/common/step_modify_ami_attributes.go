package common

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
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
		adds := make([]*ec2.LaunchPermission, len(s.Groups))
		addGroups := &ec2.ModifyImageAttributeInput{
			LaunchPermission: &ec2.LaunchPermissionModifications{},
		}

		for i, g := range s.Groups {
			groups[i] = aws.String(g)
			adds[i] = &ec2.LaunchPermission{
				Group: aws.String(g),
			}
		}
		addGroups.UserGroups = groups
		addGroups.LaunchPermission.Add = adds

		options["groups"] = addGroups
	}

	if len(s.Users) > 0 {
		users := make([]*string, len(s.Users))
		adds := make([]*ec2.LaunchPermission, len(s.Users))
		for i, u := range s.Users {
			users[i] = aws.String(u)
			adds[i] = &ec2.LaunchPermission{UserId: aws.String(u)}
		}
		options["users"] = &ec2.ModifyImageAttributeInput{
			UserIds: users,
			LaunchPermission: &ec2.LaunchPermissionModifications{
				Add: adds,
			},
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
		awsConfig := aws.Config{
			Credentials: ec2conn.Config.Credentials,
			Region:      aws.String(region),
		}
		session := session.New(&awsConfig)
		regionconn := ec2.New(session)
		for name, input := range options {
			ui.Message(fmt.Sprintf("Modifying: %s", name))
			input.ImageId = &ami
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
