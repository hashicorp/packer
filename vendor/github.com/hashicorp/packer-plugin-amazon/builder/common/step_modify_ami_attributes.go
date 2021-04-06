package common

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type StepModifyAMIAttributes struct {
	AMISkipCreateImage bool

	Users          []string
	Groups         []string
	SnapshotUsers  []string
	SnapshotGroups []string
	ProductCodes   []string
	Description    string
	Ctx            interpolate.Context

	GeneratedData *packerbuilderdata.GeneratedData
}

func (s *StepModifyAMIAttributes) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	session := state.Get("awsSession").(*session.Session)
	ui := state.Get("ui").(packersdk.Ui)

	if s.AMISkipCreateImage {
		ui.Say("Skipping AMI modify attributes...")
		return multistep.ActionContinue
	}

	amis := state.Get("amis").(map[string]string)
	snapshots := state.Get("snapshots").(map[string][]string)

	// Determine if there is any work to do.
	valid := false
	valid = valid || s.Description != ""
	valid = valid || (s.Users != nil && len(s.Users) > 0)
	valid = valid || (s.Groups != nil && len(s.Groups) > 0)
	valid = valid || (s.ProductCodes != nil && len(s.ProductCodes) > 0)
	valid = valid || (s.SnapshotUsers != nil && len(s.SnapshotUsers) > 0)
	valid = valid || (s.SnapshotGroups != nil && len(s.SnapshotGroups) > 0)

	if !valid {
		return multistep.ActionContinue
	}

	var err error
	s.Ctx.Data = extractBuildInfo(*ec2conn.Config.Region, state, s.GeneratedData)
	s.Description, err = interpolate.Render(s.Description, &s.Ctx)
	if err != nil {
		err = fmt.Errorf("Error interpolating AMI description: %s", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Construct the modify image and snapshot attribute requests we're going
	// to make. We need to make each separately since the EC2 API only allows
	// changing one type at a kind currently.
	options := make(map[string]*ec2.ModifyImageAttributeInput)
	if s.Description != "" {
		options["description"] = &ec2.ModifyImageAttributeInput{
			Description: &ec2.AttributeValue{Value: &s.Description},
		}
	}
	snapshotOptions := make(map[string]*ec2.ModifySnapshotAttributeInput)

	if len(s.Groups) > 0 {
		groups := make([]*string, len(s.Groups))
		addsImage := make([]*ec2.LaunchPermission, len(s.Groups))
		addGroups := &ec2.ModifyImageAttributeInput{
			LaunchPermission: &ec2.LaunchPermissionModifications{},
		}

		for i, g := range s.Groups {
			groups[i] = aws.String(g)
			addsImage[i] = &ec2.LaunchPermission{
				Group: aws.String(g),
			}
		}

		addGroups.UserGroups = groups
		addGroups.LaunchPermission.Add = addsImage
		options["groups"] = addGroups
	}

	if len(s.SnapshotGroups) > 0 {
		groups := make([]*string, len(s.SnapshotGroups))
		addsSnapshot := make([]*ec2.CreateVolumePermission, len(s.SnapshotGroups))
		addSnapshotGroups := &ec2.ModifySnapshotAttributeInput{
			CreateVolumePermission: &ec2.CreateVolumePermissionModifications{},
		}

		for i, g := range s.SnapshotGroups {
			groups[i] = aws.String(g)
			addsSnapshot[i] = &ec2.CreateVolumePermission{
				Group: aws.String(g),
			}
		}
		addSnapshotGroups.GroupNames = groups
		addSnapshotGroups.CreateVolumePermission.Add = addsSnapshot
		snapshotOptions["groups"] = addSnapshotGroups
	}

	if len(s.Users) > 0 {
		users := make([]*string, len(s.Users))
		addsImage := make([]*ec2.LaunchPermission, len(s.Users))
		for i, u := range s.Users {
			users[i] = aws.String(u)
			addsImage[i] = &ec2.LaunchPermission{UserId: aws.String(u)}
		}

		options["users"] = &ec2.ModifyImageAttributeInput{
			UserIds: users,
			LaunchPermission: &ec2.LaunchPermissionModifications{
				Add: addsImage,
			},
		}
	}

	if len(s.SnapshotUsers) > 0 {
		users := make([]*string, len(s.SnapshotUsers))
		addsSnapshot := make([]*ec2.CreateVolumePermission, len(s.SnapshotUsers))
		for i, u := range s.SnapshotUsers {
			users[i] = aws.String(u)
			addsSnapshot[i] = &ec2.CreateVolumePermission{UserId: aws.String(u)}
		}

		snapshotOptions["users"] = &ec2.ModifySnapshotAttributeInput{
			UserIds: users,
			CreateVolumePermission: &ec2.CreateVolumePermissionModifications{
				Add: addsSnapshot,
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

	// Modifying image attributes
	for region, ami := range amis {
		ui.Say(fmt.Sprintf("Modifying attributes on AMI (%s)...", ami))
		regionConn := ec2.New(session, &aws.Config{
			Region: aws.String(region),
		})
		for name, input := range options {
			ui.Message(fmt.Sprintf("Modifying: %s", name))
			input.ImageId = &ami
			_, err := regionConn.ModifyImageAttribute(input)
			if err != nil {
				err := fmt.Errorf("Error modify AMI attributes: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	// Modifying snapshot attributes
	for region, region_snapshots := range snapshots {
		for _, snapshot := range region_snapshots {
			ui.Say(fmt.Sprintf("Modifying attributes on snapshot (%s)...", snapshot))
			regionConn := ec2.New(session, &aws.Config{
				Region: aws.String(region),
			})
			for name, input := range snapshotOptions {
				ui.Message(fmt.Sprintf("Modifying: %s", name))
				input.SnapshotId = &snapshot
				_, err := regionConn.ModifySnapshotAttribute(input)
				if err != nil {
					err := fmt.Errorf("Error modify snapshot attributes: %s", err)
					state.Put("error", err)
					ui.Error(err.Error())
					return multistep.ActionHalt
				}
			}
		}
	}

	return multistep.ActionContinue
}

func (s *StepModifyAMIAttributes) Cleanup(state multistep.StateBag) {
	// No cleanup...
}
