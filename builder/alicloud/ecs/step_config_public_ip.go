package ecs

import (
	"context"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepConfigAlicloudPublicIP struct {
	publicIPAddress string
	RegionId        string
	SSHPrivateIp    bool
}

func (s *stepConfigAlicloudPublicIP) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packersdk.Ui)
	instance := state.Get("instance").(*ecs.Instance)

	if s.SSHPrivateIp {
		ipaddress := instance.InnerIpAddress.IpAddress
		if len(ipaddress) == 0 {
			ui.Say("Failed to get private ip of instance")
			return multistep.ActionHalt
		}
		state.Put("ipaddress", ipaddress[0])
		return multistep.ActionContinue
	}

	allocatePublicIpAddressRequest := ecs.CreateAllocatePublicIpAddressRequest()
	allocatePublicIpAddressRequest.InstanceId = instance.InstanceId
	ipaddress, err := client.AllocatePublicIpAddress(allocatePublicIpAddressRequest)
	if err != nil {
		return halt(state, err, "Error allocating public ip")
	}

	s.publicIPAddress = ipaddress.IpAddress
	ui.Say(fmt.Sprintf("Allocated public ip address %s.", ipaddress.IpAddress))
	state.Put("ipaddress", ipaddress.IpAddress)
	return multistep.ActionContinue
}

func (s *stepConfigAlicloudPublicIP) Cleanup(state multistep.StateBag) {

}
