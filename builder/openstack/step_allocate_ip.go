package openstack

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/floatingip"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

type StepAllocateIp struct {
	FloatingIpPool string
	FloatingIp     string
}

func (s *StepAllocateIp) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(Config)
	server := state.Get("server").(*servers.Server)

	// We need the v2 compute client
	client, err := config.computeV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing compute client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	var instanceIp floatingip.FloatingIP

	// This is here in case we error out before putting instanceIp into the
	// statebag below, because it is requested by Cleanup()
	state.Put("access_ip", &instanceIp)

	if s.FloatingIp != "" {
		instanceIp.IP = s.FloatingIp
	} else if s.FloatingIpPool != "" {
		ui.Say(fmt.Sprintf("Creating floating IP..."))
		ui.Message(fmt.Sprintf("Pool: %s", s.FloatingIpPool))
		newIp, err := floatingip.Create(client, floatingip.CreateOpts{
			Pool: s.FloatingIpPool,
		}).Extract()
		if err != nil {
			err := fmt.Errorf("Error creating floating ip from pool '%s'", s.FloatingIpPool)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		instanceIp = *newIp
		ui.Message(fmt.Sprintf("Created floating IP: %s", instanceIp.IP))
	}

	if instanceIp.IP != "" {
		ui.Say(fmt.Sprintf("Associating floating IP with server..."))
		ui.Message(fmt.Sprintf("IP: %s", instanceIp.IP))
		err := floatingip.Associate(client, server.ID, instanceIp.IP).ExtractErr()
		if err != nil {
			err := fmt.Errorf(
				"Error associating floating IP %s with instance: %s",
				instanceIp.IP, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		ui.Message(fmt.Sprintf(
			"Added floating IP %s to instance!", instanceIp.IP))
	}

	state.Put("access_ip", &instanceIp)
	return multistep.ActionContinue
}

func (s *StepAllocateIp) Cleanup(state multistep.StateBag) {
	config := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)
	instanceIp := state.Get("access_ip").(*floatingip.FloatingIP)

	// We need the v2 compute client
	client, err := config.computeV2Client()
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error deleting temporary floating IP %s", instanceIp.IP))
		return
	}

	if s.FloatingIpPool != "" && instanceIp.ID != "" {
		if err := floatingip.Delete(client, instanceIp.ID).ExtractErr(); err != nil {
			ui.Error(fmt.Sprintf(
				"Error deleting temporary floating IP %s", instanceIp.IP))
			return
		}

		ui.Say(fmt.Sprintf("Deleted temporary floating IP %s", instanceIp.IP))
	}
}
