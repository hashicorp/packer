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

type stepConfigAlicloudVPC struct {
	VpcId     string
	CidrBlock string //192.168.0.0/16 or 172.16.0.0/16 (default)
	VpcName   string
	isCreate  bool
}

func (s *stepConfigAlicloudVPC) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(Config)
	client := state.Get("client").(*ecs.Client)
	ui := state.Get("ui").(packer.Ui)

	if len(s.VpcId) != 0 {
		vpcs, _, err := client.DescribeVpcs(&ecs.DescribeVpcsArgs{
			VpcId:    s.VpcId,
			RegionId: common.Region(config.AlicloudRegion),
		})
		if err != nil {
			ui.Say(fmt.Sprintf("Failed querying vpcs: %s", err))
			state.Put("error", err)
			return multistep.ActionHalt
		}
		if len(vpcs) > 0 {
			vpc := vpcs[0]
			state.Put("vpcid", vpc.VpcId)
			s.isCreate = false
			return multistep.ActionContinue
		}
		message := fmt.Sprintf("The specified vpc {%s} doesn't exist.", s.VpcId)
		state.Put("error", errors.New(message))
		ui.Say(message)
		return multistep.ActionHalt

	}
	ui.Say("Creating vpc")
	vpc, err := client.CreateVpc(&ecs.CreateVpcArgs{
		RegionId:  common.Region(config.AlicloudRegion),
		CidrBlock: s.CidrBlock,
		VpcName:   s.VpcName,
	})
	if err != nil {
		state.Put("error", err)
		ui.Say(fmt.Sprintf("Failed creating vpc: %s", err))
		return multistep.ActionHalt
	}
	err = client.WaitForVpcAvailable(common.Region(config.AlicloudRegion), vpc.VpcId, ALICLOUD_DEFAULT_SHORT_TIMEOUT)
	if err != nil {
		state.Put("error", err)
		ui.Say(fmt.Sprintf("Failed waiting for vpc to become available: %s", err))
		return multistep.ActionHalt
	}

	state.Put("vpcid", vpc.VpcId)
	s.isCreate = true
	s.VpcId = vpc.VpcId
	return multistep.ActionContinue
}

func (s *stepConfigAlicloudVPC) Cleanup(state multistep.StateBag) {
	if !s.isCreate {
		return
	}

	client := state.Get("client").(*ecs.Client)
	ui := state.Get("ui").(packer.Ui)

	message(state, "VPC")
	timeoutPoint := time.Now().Add(60 * time.Second)
	for {
		if err := client.DeleteVpc(s.VpcId); err != nil {
			e, _ := err.(*common.Error)
			if (e.Code == "DependencyViolation.Instance" || e.Code == "DependencyViolation.RouteEntry" ||
				e.Code == "DependencyViolation.VSwitch" ||
				e.Code == "DependencyViolation.SecurityGroup") && time.Now().Before(timeoutPoint) {
				time.Sleep(1 * time.Second)
				continue
			}
			ui.Error(fmt.Sprintf("Error deleting vpc, it may still be around: %s", err))
			return
		}
		break
	}
}
