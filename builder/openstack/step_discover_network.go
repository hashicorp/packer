package openstack

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepDiscoverNetwork struct {
	Networks              []string
	NetworkDiscoveryCIDRs []string
	Ports                 []string
}

func (s *StepDiscoverNetwork) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	networkClient, err := config.networkV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing network client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	networks := []servers.Network{}
	for _, port := range s.Ports {
		networks = append(networks, servers.Network{Port: port})
	}
	for _, uuid := range s.Networks {
		networks = append(networks, servers.Network{UUID: uuid})
	}

	cidrs := s.NetworkDiscoveryCIDRs
	if len(networks) == 0 && len(cidrs) > 0 {
		ui.Say(fmt.Sprintf("Discovering provisioning network..."))

		networkID, err := DiscoverProvisioningNetwork(networkClient, cidrs)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}

		ui.Message(fmt.Sprintf("Found network ID: %s", networkID))
		networks = append(networks, servers.Network{UUID: networkID})
	}

	state.Put("networks", networks)
	return multistep.ActionContinue
}

func (s *StepDiscoverNetwork) Cleanup(state multistep.StateBag) {}
