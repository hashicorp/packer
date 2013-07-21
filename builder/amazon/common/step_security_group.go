package common

import (
	"cgl.tideland.biz/identifier"
	"encoding/hex"
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

type StepSecurityGroup struct {
	SecurityGroupId string
	SSHPort         int

	createdGroupId string
}

func (s *StepSecurityGroup) Run(state map[string]interface{}) multistep.StepAction {
	ec2conn := state["ec2"].(*ec2.EC2)
	ui := state["ui"].(packer.Ui)

	if s.SecurityGroupId != "" {
		log.Printf("Using specified security group: %s", s.SecurityGroupId)
		state["securityGroupId"] = s.SecurityGroupId
		return multistep.ActionContinue
	}

	if s.SSHPort == 0 {
		panic("SSHPort must be set to a non-zero value.")
	}

	// Create the group
	ui.Say("Creating temporary security group for this instance...")
	groupName := fmt.Sprintf("packer %s", hex.EncodeToString(identifier.NewUUID().Raw()))
	log.Printf("Temporary group name: %s", groupName)
	groupResp, err := ec2conn.CreateSecurityGroup(ec2.SecurityGroup{Name: groupName, Description: "Temporary group for Packer", VpcId: config.VpcId})
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the group ID so we can delete it later
	s.createdGroupId = groupResp.Id

	// Authorize the SSH access
	perms := []ec2.IPPerm{
		ec2.IPPerm{
			Protocol:  "tcp",
			FromPort:  s.SSHPort,
			ToPort:    s.SSHPort,
			SourceIPs: []string{"0.0.0.0/0"},
		},
	}

	ui.Say("Authorizing SSH access on the temporary security group...")
	if _, err := ec2conn.AuthorizeSecurityGroup(groupResp.SecurityGroup, perms); err != nil {
		err := fmt.Errorf("Error creating temporary security group: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set some state data for use in future steps
	state["securityGroupId"] = s.createdGroupId

	return multistep.ActionContinue
}

func (s *StepSecurityGroup) Cleanup(state map[string]interface{}) {
	if s.createdGroupId == "" {
		return
	}

	ec2conn := state["ec2"].(*ec2.EC2)
	ui := state["ui"].(packer.Ui)

	ui.Say("Deleting temporary security group...")
	_, err := ec2conn.DeleteSecurityGroup(ec2.SecurityGroup{Id: s.createdGroupId})
	if err != nil {
		log.Printf("Error deleting security group: %s", err)
		ui.Error(fmt.Sprintf(
			"Error cleaning up security group. Please delete the group manually: %s", s.createdGroupId))
	}
}
