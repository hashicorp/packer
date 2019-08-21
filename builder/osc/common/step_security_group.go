package common

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/outscale/osc-go/oapi"
)

type StepSecurityGroup struct {
	CommConfig            *communicator.Config
	SecurityGroupFilter   SecurityGroupFilterOptions
	SecurityGroupIds      []string
	TemporarySGSourceCidr string

	createdGroupId string
}

func (s *StepSecurityGroup) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	oapiconn := state.Get("oapi").(*oapi.Client)
	ui := state.Get("ui").(packer.Ui)
	netId := state.Get("net_id").(string)

	if len(s.SecurityGroupIds) > 0 {
		resp, err := oapiconn.POST_ReadSecurityGroups(
			oapi.ReadSecurityGroupsRequest{
				Filters: oapi.FiltersSecurityGroup{
					SecurityGroupIds: s.SecurityGroupIds,
				},
			},
		)
		if err != nil || resp.OK == nil || len(resp.OK.SecurityGroups) <= 0 {
			err := fmt.Errorf("Couldn't find specified security group: %s", err)
			log.Printf("[DEBUG] %s", err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}

		log.Printf("Using specified security groups: %v", s.SecurityGroupIds)
		state.Put("securityGroupIds", s.SecurityGroupIds)
		return multistep.ActionContinue
	}

	if !s.SecurityGroupFilter.Empty() {

		params := oapi.ReadSecurityGroupsRequest{}
		if netId != "" {
			s.SecurityGroupFilter.Filters["net-id"] = netId
		}
		params.Filters = buildSecurityGroupFilters(s.SecurityGroupFilter.Filters)

		log.Printf("Using SecurityGroup Filters %v", params)

		sgResp, err := oapiconn.POST_ReadSecurityGroups(params)
		if err != nil || sgResp.OK == nil {
			err := fmt.Errorf("Couldn't find security groups for filter: %s", err)
			log.Printf("[DEBUG] %s", err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}

		securityGroupIds := []string{}
		for _, sg := range sgResp.OK.SecurityGroups {
			securityGroupIds = append(securityGroupIds, sg.SecurityGroupId)
		}

		ui.Message(fmt.Sprintf("Found Security Group(s): %s", strings.Join(securityGroupIds, ", ")))
		state.Put("securityGroupIds", securityGroupIds)

		return multistep.ActionContinue
	}

	port := s.CommConfig.Port()
	if port == 0 {
		if s.CommConfig.Type != "none" {
			panic("port must be set to a non-zero value.")
		}
	}

	// Create the group
	groupName := fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID())
	ui.Say(fmt.Sprintf("Creating temporary security group for this instance: %s", groupName))
	group := oapi.CreateSecurityGroupRequest{
		SecurityGroupName: groupName,
		Description:       "Temporary group for Packer",
	}

	group.NetId = netId

	groupResp, err := oapiconn.POST_CreateSecurityGroup(group)
	if err != nil {
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// Set the group ID so we can delete it later
	s.createdGroupId = groupResp.OK.SecurityGroup.SecurityGroupId

	// Wait for the security group become available for authorizing
	log.Printf("[DEBUG] Waiting for temporary security group: %s", s.createdGroupId)
	err = waitForSecurityGroup(oapiconn, s.createdGroupId)
	if err == nil {
		log.Printf("[DEBUG] Found security group %s", s.createdGroupId)
	} else {
		err := fmt.Errorf("Timed out waiting for security group %s: %s", s.createdGroupId, err)
		log.Printf("[DEBUG] %s", err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// Authorize the SSH access for the security group
	groupRules := oapi.CreateSecurityGroupRuleRequest{
		SecurityGroupId: groupResp.OK.SecurityGroup.SecurityGroupId,
		Flow:            "Inbound",
		Rules: []oapi.SecurityGroupRule{
			{
				FromPortRange: int64(port),
				ToPortRange:   int64(port),
				IpRanges:      []string{s.TemporarySGSourceCidr},
				IpProtocol:    "tcp",
			},
		},
	}

	ui.Say(fmt.Sprintf(
		"Authorizing access to port %d from %s in the temporary security group...",
		port, s.TemporarySGSourceCidr))
	_, err = oapiconn.POST_CreateSecurityGroupRule(groupRules)
	if err != nil {
		err := fmt.Errorf("Error authorizing temporary security group: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set some state data for use in future steps
	state.Put("securityGroupIds", []string{s.createdGroupId})

	return multistep.ActionContinue
}

func (s *StepSecurityGroup) Cleanup(state multistep.StateBag) {
	if s.createdGroupId == "" {
		return
	}

	oapiconn := state.Get("oapi").(*oapi.Client)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting temporary security group...")

	var err error
	for i := 0; i < 5; i++ {
		_, err = oapiconn.POST_DeleteSecurityGroup(oapi.DeleteSecurityGroupRequest{SecurityGroupId: s.createdGroupId})
		if err == nil {
			break
		}

		log.Printf("Error deleting security group: %s", err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up security group. Please delete the group manually: %s", s.createdGroupId))
	}
}
