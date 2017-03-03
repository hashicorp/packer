package ecs

import (
	"errors"
	"fmt"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"time"
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
			ui.Say(fmt.Sprintf("Query alicloud security grouip failed: %s", err))
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
		message := fmt.Sprintf("The specific security group {%s} isn't exist.", s.SecurityGroupId)
		state.Put("error", errors.New(message))
		ui.Say(message)
		return multistep.ActionHalt

	}
	var securityGroupId string
	ui.Say("Start creating security groups...")
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
		ui.Say(fmt.Sprintf("Create security group failed %v", err))
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
		ui.Say(fmt.Sprintf("authorzie security group failed %v", err))
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
		ui.Say(fmt.Sprintf("authorzie security group failed %v", err))
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
	start := time.Now().Add(10 * time.Second)
	for {
		if err := client.DeleteSecurityGroup(common.Region(s.RegionId), s.SecurityGroupId); err != nil {
			e, _ := err.(*common.Error)
			if e.Code == "DependencyViolation" && time.Now().Before(start) {
				time.Sleep(1 * time.Second)
				continue
			}
			ui.Error(fmt.Sprintf("Error delete security group failed, may still be around: %s", err))
			return
		}
		break
	}
}
