package ebs

import (
	"cgl.tideland.biz/identifier"
	"encoding/hex"
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

type stepSecurityGroup struct {
	groupId string
}

func (s *stepSecurityGroup) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(config)
	ec2conn := state["ec2"].(*ec2.EC2)
	ui := state["ui"].(packer.Ui)

	if config.SecurityGroupId != "" {
		log.Printf("Using specified security group: %s", config.SecurityGroupId)
		state["securityGroupId"] = config.SecurityGroupId
		return multistep.ActionContinue
	}

	// Create the group
	ui.Say("Creating temporary security group for this instance...")
	groupName := fmt.Sprintf("packer %s", hex.EncodeToString(identifier.NewUUID().Raw()))
	log.Printf("Temporary group name: %s", groupName)
	groupResp, err := ec2conn.CreateSecurityGroup(groupName, "Temporary group for Packer")
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the group ID so we can delete it later
	s.groupId = groupResp.Id

	// Authorize the SSH access
	perms := []ec2.IPPerm{
		ec2.IPPerm{
			Protocol:  "tcp",
			FromPort:  config.SSHPort,
			ToPort:    config.SSHPort,
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
	state["securityGroupId"] = s.groupId

	return multistep.ActionContinue
}

func (s *stepSecurityGroup) Cleanup(state map[string]interface{}) {
	if s.groupId == "" {
		return
	}

	ec2conn := state["ec2"].(*ec2.EC2)
	ui := state["ui"].(packer.Ui)

	ui.Say("Deleting temporary security group...")
	_, err := ec2conn.DeleteSecurityGroup(ec2.SecurityGroup{Id: s.groupId})
	if err != nil {
		log.Printf("Error deleting security group: %s", err)
		ui.Error(fmt.Sprintf(
			"Error cleaning up security group. Please delete the group manually: %s", s.groupId))
	}
}
