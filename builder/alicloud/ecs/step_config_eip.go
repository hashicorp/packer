package ecs

import (
	"fmt"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
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
	ui.Say("Allocating eip")
	ipaddress, allocateId, err := client.AllocateEipAddress(&ecs.AllocateEipAddressArgs{
		RegionId: common.Region(s.RegionId), InternetChargeType: common.InternetChargeType(s.InternetChargeType),
	})
	if err != nil {
		state.Put("error", err)
		ui.Say(fmt.Sprintf("Error allocating eip: %s", err))
		return multistep.ActionHalt
	}
	s.allocatedId = allocateId
	if err = client.WaitForEip(common.Region(s.RegionId), allocateId,
		ecs.EipStatusAvailable, ALICLOUD_DEFAULT_SHORT_TIMEOUT); err != nil {
		state.Put("error", err)
		ui.Say(fmt.Sprintf("Error allocating eip: %s", err))
		return multistep.ActionHalt
	}

	if err = client.AssociateEipAddress(allocateId, instance.InstanceId); err != nil {
		state.Put("error", err)
		ui.Say(fmt.Sprintf("Error binding eip: %s", err))
		return multistep.ActionHalt
	}

	if err = client.WaitForEip(common.Region(s.RegionId), allocateId,
		ecs.EipStatusInUse, ALICLOUD_DEFAULT_SHORT_TIMEOUT); err != nil {
		state.Put("error", err)
		ui.Say(fmt.Sprintf("Error associating eip: %s", err))
		return multistep.ActionHalt
	}
	ui.Say(fmt.Sprintf("Allocated eip %s", ipaddress))
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
		ui.Say(fmt.Sprintf("Failed to unassociate eip."))
	}

	if err := client.WaitForEip(common.Region(s.RegionId), s.allocatedId, ecs.EipStatusAvailable, ALICLOUD_DEFAULT_SHORT_TIMEOUT); err != nil {
		ui.Say(fmt.Sprintf("Timeout while unassociating eip."))
	}
	if err := client.ReleaseEipAddress(s.allocatedId); err != nil {
		ui.Say(fmt.Sprintf("Failed to release eip."))
	}

}
