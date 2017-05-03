package openstack_id3

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/floatingip"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

type StepAllocateIp struct {
	FloatingIpPool string
	FloatingIp     string
}

func (s *StepAllocateIp) Run(state multistep.StateBag) multistep.StepAction {

	ui := state.Get("ui").(packer.Ui)
	computeClient := state.Get("compute_client").(*gophercloud.ServiceClient)
	server := state.Get("server").(*servers.Server)

	var instanceIp floatingip.FloatingIP
	// This is here in case we error out before putting instanceIp into the
	// statebag below, because it is requested by Cleanup()
	state.Put("access_ip", instanceIp)

	if s.FloatingIp != "" {
		instanceIp.IP = s.FloatingIp
	} else if s.FloatingIpPool != "" {
		// Obtain new floating IP. Supports both networks pool id and name
		newIp, err := floatingip.Create(computeClient, floatingip.CreateOpts{Pool: s.FloatingIpPool}).Extract()
		if err != nil {
			err := fmt.Errorf("Error creating floating ip from pool '%s'", s.FloatingIpPool)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		instanceIp = *newIp
		ui.Say(fmt.Sprintf("Created temporary floating IP %s...", instanceIp.IP))
	}
	if instanceIp.IP != "" {
		err := floatingip.Associate(computeClient, server.ID, instanceIp.IP).ExtractErr()
		if err != nil {
			err := fmt.Errorf("Error associating floating IP %s with instance.", instanceIp.IP)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		} else {
			ui.Say(fmt.Sprintf("Added floating IP %s to instance...", instanceIp.IP))
		}
	}
	state.Put("access_ip", instanceIp)
	return multistep.ActionContinue
}

func (s *StepAllocateIp) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	computeClient := state.Get("compute_client").(*gophercloud.ServiceClient)
	instanceIp := state.Get("access_ip").(floatingip.FloatingIP)

	if s.FloatingIpPool != "" && instanceIp.ID != "" {
		err := floatingip.Delete(computeClient, instanceIp.ID).ExtractErr()
		if err != nil {
			ui.Error(fmt.Sprintf("Error deleting temporary floating IP %s", instanceIp.IP))
			return
		}
		ui.Say(fmt.Sprintf("Deleted temporary floating IP %s", instanceIp.IP))
	}
}
