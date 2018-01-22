package ecs

import (
	"fmt"

	"github.com/denverdino/aliyungo/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepConfigAlicloudPublicIP struct {
	publicIPAdress string
	RegionId       string
}

func (s *stepConfigAlicloudPublicIP) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ecs.Client)
	ui := state.Get("ui").(packer.Ui)
	instance := state.Get("instance").(*ecs.InstanceAttributesType)

	ipaddress, err := client.AllocatePublicIpAddress(instance.InstanceId)
	if err != nil {
		state.Put("error", err)
		ui.Say(fmt.Sprintf("Error allocating public ip: %s", err))
		return multistep.ActionHalt
	}
	s.publicIPAdress = ipaddress
	ui.Say(fmt.Sprintf("Allocated public ip address %s.", ipaddress))
	state.Put("ipaddress", ipaddress)
	return multistep.ActionContinue
}

func (s *stepConfigAlicloudPublicIP) Cleanup(state multistep.StateBag) {

}
