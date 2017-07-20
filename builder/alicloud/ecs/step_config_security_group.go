package ecs

import (
	"errors"
	"fmt"
	"time"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type stepConfigAlicloudSecurityGroup struct {
	SecurityGroupId   string
	SecurityGroupName string
	Description       string
	VpcId             string
	RegionId          string
	isCreate          bool
}

func (s *stepConfigAlicloudSecurityGroup) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ecs.Client)
	ui := state.Get("ui").(packer.Ui)
	networkType := state.Get("networktype").(InstanceNetWork)

	var securityGroupItems []ecs.SecurityGroupItemType
	var err error
	if len(s.SecurityGroupId) != 0 {
		if networkType == VpcNet {
			vpcId := state.Get("vpcid").(string)
			securityGroupItems, _, err = client.DescribeSecurityGroups(&ecs.DescribeSecurityGroupsArgs{
				VpcId:    vpcId,
				RegionId: common.Region(s.RegionId),
			})
		} else {
			securityGroupItems, _, err = client.DescribeSecurityGroups(&ecs.DescribeSecurityGroupsArgs{
				RegionId: common.Region(s.RegionId),
			})
		}

		if err != nil {
			ui.Say(fmt.Sprintf("Failed querying security group: %s", err))
			state.Put("error", err)
			return multistep.ActionHalt
		}
		for _, securityGroupItem := range securityGroupItems {
			if securityGroupItem.SecurityGroupId == s.SecurityGroupId {
				state.Put("securitygroupid", s.SecurityGroupId)
				s.isCreate = false
				return multistep.ActionContinue
			}
		}
		s.isCreate = false
		message := fmt.Sprintf("The specified security group {%s} doesn't exist.", s.SecurityGroupId)
		state.Put("error", errors.New(message))
		ui.Say(message)
		return multistep.ActionHalt

	}
	var securityGroupId string
	ui.Say("Creating security groups...")
	if networkType == VpcNet {
		vpcId := state.Get("vpcid").(string)
		securityGroupId, err = client.CreateSecurityGroup(&ecs.CreateSecurityGroupArgs{
			RegionId:          common.Region(s.RegionId),
			SecurityGroupName: s.SecurityGroupName,
			VpcId:             vpcId,
		})
	} else {
		securityGroupId, err = client.CreateSecurityGroup(&ecs.CreateSecurityGroupArgs{
			RegionId:          common.Region(s.RegionId),
			SecurityGroupName: s.SecurityGroupName,
		})
	}
	if err != nil {
		state.Put("error", err)
		ui.Say(fmt.Sprintf("Failed creating security group %s.", err))
		return multistep.ActionHalt
	}
	state.Put("securitygroupid", securityGroupId)
	s.isCreate = true
	s.SecurityGroupId = securityGroupId
	err = client.AuthorizeSecurityGroupEgress(&ecs.AuthorizeSecurityGroupEgressArgs{
		SecurityGroupId: securityGroupId,
		RegionId:        common.Region(s.RegionId),
		IpProtocol:      ecs.IpProtocolAll,
		PortRange:       "-1/-1",
		NicType:         ecs.NicTypeInternet,
		DestCidrIp:      "0.0.0.0/0", //The input parameter "DestGroupId" or "DestCidrIp" cannot be both blank.
	})
	if err != nil {
		state.Put("error", err)
		ui.Say(fmt.Sprintf("Failed authorizing security group: %s", err))
		return multistep.ActionHalt
	}
	err = client.AuthorizeSecurityGroup(&ecs.AuthorizeSecurityGroupArgs{
		SecurityGroupId: securityGroupId,
		RegionId:        common.Region(s.RegionId),
		IpProtocol:      ecs.IpProtocolAll,
		PortRange:       "-1/-1",
		NicType:         ecs.NicTypeInternet,
		SourceCidrIp:    "0.0.0.0/0", //The input parameter "SourceGroupId" or "SourceCidrIp" cannot be both blank.
	})
	if err != nil {
		state.Put("error", err)
		ui.Say(fmt.Sprintf("Failed authorizing security group: %s", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepConfigAlicloudSecurityGroup) Cleanup(state multistep.StateBag) {
	if !s.isCreate {
		return
	}

	client := state.Get("client").(*ecs.Client)
	ui := state.Get("ui").(packer.Ui)

	message(state, "security group")
	start := time.Now().Add(120 * time.Second)
	for {
		if err := client.DeleteSecurityGroup(common.Region(s.RegionId), s.SecurityGroupId); err != nil {
			e, _ := err.(*common.Error)
			if e.Code == "DependencyViolation" && time.Now().Before(start) {
				time.Sleep(5 * time.Second)
				continue
			}
			ui.Error(fmt.Sprintf("Failed to delete security group, it may still be around: %s", err))
			return
		}
		break
	}
}
