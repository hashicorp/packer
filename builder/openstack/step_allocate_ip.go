package openstack

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud"
)

type StepAllocateIp struct {
	FloatingIpPool string
	FloatingIp     string
}

func (s *StepAllocateIp) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	csp := state.Get("csp").(gophercloud.CloudServersProvider)
	server := state.Get("server").(*gophercloud.Server)

	var instanceIp gophercloud.FloatingIp
	// This is here in case we error out before putting instanceIp into the
	// statebag below, because it is requested by Cleanup()
	state.Put("access_ip", instanceIp)

	if s.FloatingIp != "" {
		instanceIp.Ip = s.FloatingIp
	} else if s.FloatingIpPool != "" {
		newIp, err := csp.CreateFloatingIp(s.FloatingIpPool)
		if err != nil {
			err := fmt.Errorf("Error creating floating ip from pool '%s'", s.FloatingIpPool)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		instanceIp = newIp
		ui.Say(fmt.Sprintf("Created temporary floating IP %s...", instanceIp.Ip))
	}

	if instanceIp.Ip != "" {
		if err := csp.AssociateFloatingIp(server.Id, instanceIp); err != nil {
			err := fmt.Errorf("Error associating floating IP %s with instance.", instanceIp.Ip)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		} else {
			ui.Say(fmt.Sprintf("Added floating IP %s to instance...", instanceIp.Ip))
		}
	}

	state.Put("access_ip", instanceIp)

	return multistep.ActionContinue
}

func (s *StepAllocateIp) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	csp := state.Get("csp").(gophercloud.CloudServersProvider)
	instanceIp := state.Get("access_ip").(gophercloud.FloatingIp)
	if s.FloatingIpPool != "" && instanceIp.Id != 0 {
		if err := csp.DeleteFloatingIp(instanceIp); err != nil {
			ui.Error(fmt.Sprintf("Error deleting temporary floating IP %s", instanceIp.Ip))
			return
		}
		ui.Say(fmt.Sprintf("Deleted temporary floating IP %s", instanceIp.Ip))
	}
}
