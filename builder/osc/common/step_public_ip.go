package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/outscale/osc-go/oapi"
)

type StepPublicIp struct {
	AssociatePublicIpAddress bool
	Comm                     *communicator.Config
	publicIpId               string
	Debug                    bool

	doCleanup bool
}

func (s *StepPublicIp) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	oapiconn := state.Get("oapi").(*oapi.Client)
	netId := state.Get("net_id").(string)
	subnetId := state.Get("subnet_id").(string)

	if netId == "" || subnetId == "" || !s.AssociatePublicIpAddress {
		// In this case, we are in the public Cloud, so we'll
		// not explicitely allocate a public IP.
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Creating temporary PublicIp for instance in subnet %s (net %s)", subnetId, netId))

	publicIpResp, err := oapiconn.POST_CreatePublicIp(oapi.CreatePublicIpRequest{})
	if err != nil {
		state.Put("error", fmt.Errorf("Error creating temporary PublicIp: %s", err))
		return multistep.ActionHalt
	}

	// From there, we have a Public Ip to destroy.
	s.doCleanup = true

	// Set some data for use in future steps
	s.publicIpId = publicIpResp.OK.PublicIp.PublicIpId
	state.Put("publicip_id", publicIpResp.OK.PublicIp.PublicIpId)

	return multistep.ActionContinue
}

func (s *StepPublicIp) Cleanup(state multistep.StateBag) {
	if !s.doCleanup {
		return
	}

	oapiconn := state.Get("oapi").(*oapi.Client)
	ui := state.Get("ui").(packer.Ui)

	// Remove the Public IP
	ui.Say("Deleting temporary PublicIp...")
	_, err := oapiconn.POST_DeletePublicIp(oapi.DeletePublicIpRequest{PublicIpId: s.publicIpId})
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up PublicIp. Please delete the PublicIp manually: %s", s.publicIpId))
	}

}
