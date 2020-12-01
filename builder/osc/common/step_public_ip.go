package common

import (
	"context"
	"fmt"

	"github.com/antihax/optional"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/outscale/osc-sdk-go/osc"
)

type StepPublicIp struct {
	AssociatePublicIpAddress bool
	Comm                     *communicator.Config
	publicIpId               string
	Debug                    bool

	doCleanup bool
}

func (s *StepPublicIp) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	var (
		ui   = state.Get("ui").(packersdk.Ui)
		conn = state.Get("osc").(*osc.APIClient)
	)

	if !s.AssociatePublicIpAddress {

		// In this case, we are in the public Cloud, so we'll
		// not explicitely allocate a public IP.
		return multistep.ActionContinue
	}

	ui.Say("Creating temporary PublicIp for instance ")

	resp, _, err := conn.PublicIpApi.CreatePublicIp(context.Background(), &osc.CreatePublicIpOpts{
		CreatePublicIpRequest: optional.NewInterface(osc.CreatePublicIpRequest{}),
	})

	if err != nil {
		state.Put("error", fmt.Errorf("Error creating temporary PublicIp: %s", err))
		return multistep.ActionHalt
	}

	// From there, we have a Public Ip to destroy.
	s.doCleanup = true

	// Set some data for use in future steps
	s.publicIpId = resp.PublicIp.PublicIpId
	state.Put("publicip_id", resp.PublicIp.PublicIpId)

	return multistep.ActionContinue
}

func (s *StepPublicIp) Cleanup(state multistep.StateBag) {
	if !s.doCleanup {
		return
	}

	var (
		conn = state.Get("osc").(*osc.APIClient)
		ui   = state.Get("ui").(packersdk.Ui)
	)

	// Remove the Public IP
	ui.Say("Deleting temporary PublicIp...")
	_, _, err := conn.PublicIpApi.DeletePublicIp(context.Background(), &osc.DeletePublicIpOpts{
		DeletePublicIpRequest: optional.NewInterface(osc.DeletePublicIpRequest{
			PublicIpId: s.publicIpId,
		}),
	})

	if err != nil {
		ui.Error(fmt.Sprintf("Error cleaning up PublicIp. Please delete the PublicIp manually: %s", s.publicIpId))
	}
}
