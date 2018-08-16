package openstack

import (
	"context"
	"fmt"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepAllocateIp struct {
	FloatingIPNetwork string
	FloatingIP        string
	ReuseIPs          bool
}

func (s *StepAllocateIp) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(Config)
	server := state.Get("server").(*servers.Server)

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

	var instanceIP floatingips.FloatingIP

	// This is here in case we error out before putting instanceIp into the
	// statebag below, because it is requested by Cleanup()
	state.Put("access_ip", &instanceIP)

	// Try to use floating IP provided by the user or find a free floating IP.
	if s.FloatingIP != "" {
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
	}

	// Create a new floating IP if it wasn't obtained in the previous step.
	if instanceIP.ID == "" {
		var floatingNetwork string

		if s.FloatingIPNetwork != "" {
			// Validate provided external network reference and get an ID.
			floatingNetwork, err = CheckExternalNetworkRef(networkClient, s.FloatingIPNetwork)
			if err != nil {
				err := fmt.Errorf("Error using the provided floating_ip_network: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		} else {
			// Search for the external network that can be used for the floating IPs.
			ui.Say(fmt.Sprintf("Searching for the external network..."))
			externalNetwork, err := FindExternalNetwork(networkClient)
			if err != nil {
				err := fmt.Errorf("Error searching the external network: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
			floatingNetwork = externalNetwork.ID
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

	// Assoctate a floating IP that was obtained in the previous steps.
	if instanceIP.ID != "" {
		ui.Say(fmt.Sprintf("Associating floating IP '%s' (%s) with instance port...",
			instanceIP.ID, instanceIP.FloatingIP))

		portID, err := GetInstancePortID(computeClient, server.ID)
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
	config := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)
	instanceIP := state.Get("access_ip").(*floatingips.FloatingIP)

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
