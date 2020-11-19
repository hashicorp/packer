package common

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/uuid"
)

type StepIamInstanceProfile struct {
	IamInstanceProfile                        string
	SkipProfileValidation                     bool
	TemporaryIamInstanceProfilePolicyDocument *PolicyDocument
	createdInstanceProfileName                string
	createdRoleName                           string
	createdPolicyName                         string
	roleIsAttached                            bool
}

func (s *StepIamInstanceProfile) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	iamsvc := state.Get("iam").(*iam.IAM)
	ui := state.Get("ui").(packersdk.Ui)

	state.Put("iamInstanceProfile", "")

	if len(s.IamInstanceProfile) > 0 {
		if !s.SkipProfileValidation {
			_, err := iamsvc.GetInstanceProfile(
				&iam.GetInstanceProfileInput{
					InstanceProfileName: aws.String(s.IamInstanceProfile),
				},
			)
			if err != nil {
				err := fmt.Errorf("Couldn't find specified instance profile: %s", err)
				log.Printf("[DEBUG] %s", err.Error())
				state.Put("error", err)
				return multistep.ActionHalt
			}
		}
		log.Printf("Using specified instance profile: %v", s.IamInstanceProfile)
		state.Put("iamInstanceProfile", s.IamInstanceProfile)
		return multistep.ActionContinue
	}

	if s.TemporaryIamInstanceProfilePolicyDocument != nil {
		// Create the profile
		profileName := fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())

		policy, err := json.Marshal(s.TemporaryIamInstanceProfilePolicyDocument)
		if err != nil {
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}

		ui.Say(fmt.Sprintf("Creating temporary instance profile for this instance: %s", profileName))

		profileResp, err := iamsvc.CreateInstanceProfile(&iam.CreateInstanceProfileInput{
			InstanceProfileName: aws.String(profileName),
		})
		if err != nil {
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
		s.createdInstanceProfileName = aws.StringValue(profileResp.InstanceProfile.InstanceProfileName)

		log.Printf("[DEBUG] Waiting for temporary instance profile: %s", s.createdInstanceProfileName)
		err = iamsvc.WaitUntilInstanceProfileExists(&iam.GetInstanceProfileInput{
			InstanceProfileName: aws.String(s.createdInstanceProfileName),
		})

		if err == nil {
			log.Printf("[DEBUG] Found instance profile %s", s.createdInstanceProfileName)
		} else {
			err := fmt.Errorf("Timed out waiting for instance profile %s: %s", s.createdInstanceProfileName, err)
			log.Printf("[DEBUG] %s", err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}

		ui.Say(fmt.Sprintf("Creating temporary role for this instance: %s", profileName))

		roleResp, err := iamsvc.CreateRole(&iam.CreateRoleInput{
			RoleName:                 aws.String(profileName),
			Description:              aws.String("Temporary role for Packer"),
			AssumeRolePolicyDocument: aws.String("{\"Version\": \"2012-10-17\",\"Statement\": [{\"Effect\": \"Allow\",\"Principal\": {\"Service\": \"ec2.amazonaws.com\"},\"Action\": \"sts:AssumeRole\"}]}"),
		})
		if err != nil {
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}

		s.createdRoleName = aws.StringValue(roleResp.Role.RoleName)

		log.Printf("[DEBUG] Waiting for temporary role: %s", s.createdInstanceProfileName)
		err = iamsvc.WaitUntilRoleExists(&iam.GetRoleInput{
			RoleName: aws.String(s.createdRoleName),
		})
		if err == nil {
			log.Printf("[DEBUG] Found temporary role %s", s.createdRoleName)
		} else {
			err := fmt.Errorf("Timed out waiting for temporary role %s: %s", s.createdRoleName, err)
			log.Printf("[DEBUG] %s", err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}

		ui.Say(fmt.Sprintf("Attaching policy to the temporary role: %s", profileName))

		_, err = iamsvc.PutRolePolicy(&iam.PutRolePolicyInput{
			RoleName:       roleResp.Role.RoleName,
			PolicyName:     aws.String(profileName),
			PolicyDocument: aws.String(string(policy)),
		})
		if err != nil {
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}

		s.createdPolicyName = aws.StringValue(roleResp.Role.RoleName)

		_, err = iamsvc.AddRoleToInstanceProfile(&iam.AddRoleToInstanceProfileInput{
			RoleName:            roleResp.Role.RoleName,
			InstanceProfileName: profileResp.InstanceProfile.InstanceProfileName,
		})
		if err != nil {
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}

		s.roleIsAttached = true
		state.Put("iamInstanceProfile", aws.StringValue(profileResp.InstanceProfile.InstanceProfileName))
	}

	return multistep.ActionContinue
}

func (s *StepIamInstanceProfile) Cleanup(state multistep.StateBag) {
	iamsvc := state.Get("iam").(*iam.IAM)
	ui := state.Get("ui").(packersdk.Ui)
	var err error

	if s.roleIsAttached == true {
		ui.Say("Detaching temporary role from instance profile...")

		_, err := iamsvc.RemoveRoleFromInstanceProfile(&iam.RemoveRoleFromInstanceProfileInput{
			InstanceProfileName: aws.String(s.createdInstanceProfileName),
			RoleName:            aws.String(s.createdRoleName),
		})
		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error %s. Please delete the role manually: %s", err.Error(), s.createdRoleName))
		}
	}

	if s.createdPolicyName != "" {
		ui.Say("Removing policy from temporary role...")
		iamsvc.DeleteRolePolicy(&iam.DeleteRolePolicyInput{
			PolicyName: aws.String(s.createdPolicyName),
			RoleName:   aws.String(s.createdRoleName),
		})
	}
	if s.createdRoleName != "" {
		ui.Say("Deleting temporary role...")

		_, err = iamsvc.DeleteRole(&iam.DeleteRoleInput{RoleName: &s.createdRoleName})
		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error %s. Please delete the role manually: %s", err.Error(), s.createdRoleName))
		}
	}

	if s.createdInstanceProfileName != "" {
		ui.Say("Deleting temporary instance profile...")

		_, err = iamsvc.DeleteInstanceProfile(&iam.DeleteInstanceProfileInput{
			InstanceProfileName: &s.createdInstanceProfileName})

		if err != nil {
			ui.Error(fmt.Sprintf(
				"Error %s. Please delete the instance profile manually: %s", err.Error(), s.createdInstanceProfileName))
		}
	}
}
