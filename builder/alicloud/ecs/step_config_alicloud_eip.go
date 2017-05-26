package ecs

import (
	"fmt"
	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type setpConfigAlicloudEIP struct {
	AssociatePublicIpAddress bool
	RegionId                 string
	InternetChargeType       string
	allocatedId              string
}

func (s *setpConfigAlicloudEIP) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ecs.Client)
	ui := state.Get("ui").(packer.Ui)
	instance := state.Get("instance").(*ecs.InstanceAttributesType)
	ui.Say("Start allocated alicloud eip")
	ipaddress, allocateId, err := client.AllocateEipAddress(&ecs.AllocateEipAddressArgs{
		RegionId: common.Region(s.RegionId), InternetChargeType: common.InternetChargeType(s.InternetChargeType),
	})
	if err != nil {
		state.Put("error", err)
		ui.Say(fmt.Sprintf("Error allocate eip: %s", err))
		return multistep.ActionHalt
	}
	s.allocatedId = allocateId
	if err = client.WaitForEip(common.Region(s.RegionId), allocateId,
		ecs.EipStatusAvailable, ALICLOUD_DEFAULT_SHORT_TIMEOUT); err != nil {
		state.Put("error", err)
		ui.Say(fmt.Sprintf("Error allocate alicloud eip: %s", err))
		return multistep.ActionHalt
	}

	if err = client.AssociateEipAddress(allocateId, instance.InstanceId); err != nil {
		state.Put("error", err)
		ui.Say(fmt.Sprintf("Error binding alicloud eip: %s", err))
		return multistep.ActionHalt
	}

	if err = client.WaitForEip(common.Region(s.RegionId), allocateId,
		ecs.EipStatusInUse, ALICLOUD_DEFAULT_SHORT_TIMEOUT); err != nil {
		state.Put("error", err)
		ui.Say(fmt.Sprintf("Error associating alicloud eip: %s", err))
		return multistep.ActionHalt
	}
	ui.Say(fmt.Sprintf("Allocated alicloud eip %s", ipaddress))
	state.Put("ipaddress", ipaddress)
	return multistep.ActionContinue
}

func (s *setpConfigAlicloudEIP) Cleanup(state multistep.StateBag) {
	if len(s.allocatedId) == 0 {
		return
	}

	client := state.Get("client").(*ecs.Client)
	instance := state.Get("instance").(*ecs.InstanceAttributesType)
	ui := state.Get("ui").(packer.Ui)

	message(state, "EIP")

	if err := client.UnassociateEipAddress(s.allocatedId, instance.InstanceId); err != nil {
		ui.Say(fmt.Sprintf("Unassociate alicloud eip failed "))
	}

	if err := client.WaitForEip(common.Region(s.RegionId), s.allocatedId,
		ecs.EipStatusAvailable, ALICLOUD_DEFAULT_SHORT_TIMEOUT); err != nil {
		ui.Say(fmt.Sprintf("Unassociate alicloud eip timeout "))
	}
	if err := client.ReleaseEipAddress(s.allocatedId); err != nil {
		ui.Say(fmt.Sprintf("Release alicloud eip failed "))
	}

}
