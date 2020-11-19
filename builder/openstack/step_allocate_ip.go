package openstack

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepAllocateIp struct {
	FloatingIPNetwork     string
	FloatingIP            string
	ReuseIPs              bool
	InstanceFloatingIPNet string
}

func (s *StepAllocateIp) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	config := state.Get("config").(*Config)
	server := state.Get("server").(*servers.Server)

	var instanceIP floatingips.FloatingIP

	// This is here in case we error out before putting instanceIp into the
	// statebag below, because it is requested by Cleanup()
	state.Put("access_ip", &instanceIP)

	if s.FloatingIP == "" && !s.ReuseIPs && s.FloatingIPNetwork == "" {
		ui.Message("Floating IP not required")
		return multistep.ActionContinue
	}

	// We need the v2 compute client
	computeClient, err := config.computeV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing compute client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// We need the v2 network client
	networkClient, err := config.networkV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing network client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// Try to Use the OpenStack floating IP by checking provided parameters in
	// the following order:
	//  - try to use "FloatingIP" ID directly if it's provided
	//  - try to find free floating IP in the project if "ReuseIPs" is set
	//  - create a new floating IP if "FloatingIPNetwork" is provided (it can be
	//    ID or name of the network).
	if s.FloatingIP != "" {
		// Try to use FloatingIP if it was provided by the user.
		freeFloatingIP, err := CheckFloatingIP(networkClient, s.FloatingIP)
		if err != nil {
			err := fmt.Errorf("Error using provided floating IP '%s': %s", s.FloatingIP, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		instanceIP = *freeFloatingIP
		ui.Message(fmt.Sprintf("Selected floating IP: '%s' (%s)", instanceIP.ID, instanceIP.FloatingIP))
		state.Put("floatingip_istemp", false)
	} else if s.ReuseIPs {
		// If ReuseIPs is set to true and we have a free floating IP, use it rather
		// than creating one.
		ui.Say(fmt.Sprint("Searching for unassociated floating IP"))
		freeFloatingIP, err := FindFreeFloatingIP(networkClient)
		if err != nil {
			err := fmt.Errorf("Error searching for floating IP: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		instanceIP = *freeFloatingIP
		ui.Message(fmt.Sprintf("Selected floating IP: '%s' (%s)", instanceIP.ID, instanceIP.FloatingIP))
		state.Put("floatingip_istemp", false)
	} else if s.FloatingIPNetwork != "" {
		// Lastly, if FloatingIPNetwork was provided by the user, we need to use it
		// to allocate a new floating IP and associate it to the instance.
		floatingNetwork, err := CheckFloatingIPNetwork(networkClient, s.FloatingIPNetwork)
		if err != nil {
			err := fmt.Errorf("Error using the provided floating_ip_network: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		ui.Say(fmt.Sprintf("Creating floating IP using network %s ...", floatingNetwork))
		newIP, err := floatingips.Create(networkClient, floatingips.CreateOpts{
			FloatingNetworkID: floatingNetwork,
		}).Extract()
		if err != nil {
			err := fmt.Errorf("Error creating floating IP from floating network '%s': %s", floatingNetwork, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		instanceIP = *newIP
		ui.Message(fmt.Sprintf("Created floating IP: '%s' (%s)", instanceIP.ID, instanceIP.FloatingIP))
		state.Put("floatingip_istemp", true)
	}

	// Assoctate a floating IP if it was obtained in the previous steps.
	if instanceIP.ID != "" {
		ui.Say(fmt.Sprintf("Associating floating IP '%s' (%s) with instance port...",
			instanceIP.ID, instanceIP.FloatingIP))

		portID, err := GetInstancePortID(computeClient, server.ID, s.InstanceFloatingIPNet)
		if err != nil {
			err := fmt.Errorf("Error getting interfaces of the instance '%s': %s", server.ID, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		_, err = floatingips.Update(networkClient, instanceIP.ID, floatingips.UpdateOpts{
			PortID: &portID,
		}).Extract()
		if err != nil {
			err := fmt.Errorf(
				"Error associating floating IP '%s' (%s) with instance port '%s': %s",
				instanceIP.ID, instanceIP.FloatingIP, portID, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		ui.Message(fmt.Sprintf(
			"Added floating IP '%s' (%s) to instance!", instanceIP.ID, instanceIP.FloatingIP))
	}

	state.Put("access_ip", &instanceIP)
	return multistep.ActionContinue
}

func (s *StepAllocateIp) Cleanup(state multistep.StateBag) {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)
	instanceIP := state.Get("access_ip").(*floatingips.FloatingIP)

	// Don't clean up if unless required
	if instanceIP.ID == "" && instanceIP.FloatingIP == "" {
		return
	}

	// Don't delete pool addresses we didn't allocate
	if state.Get("floatingip_istemp") == false {
		return
	}

	// We need the v2 network client
	client, err := config.networkV2Client()
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error deleting temporary floating IP '%s' (%s)", instanceIP.ID, instanceIP.FloatingIP))
		return
	}

	if instanceIP.ID != "" {
		if err := floatingips.Delete(client, instanceIP.ID).ExtractErr(); err != nil {
			ui.Error(fmt.Sprintf(
				"Error deleting temporary floating IP '%s' (%s)", instanceIP.ID, instanceIP.FloatingIP))
			return
		}

		ui.Say(fmt.Sprintf("Deleted temporary floating IP '%s' (%s)", instanceIP.ID, instanceIP.FloatingIP))
	}
}
