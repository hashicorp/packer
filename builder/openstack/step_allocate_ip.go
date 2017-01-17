package openstack

import (
	"fmt"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/floatingips"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
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

	var instanceIp floatingips.FloatingIP

	// This is here in case we error out before putting instanceIp into the
	// statebag below, because it is requested by Cleanup()
	state.Put("access_ip", &instanceIp)

	if s.FloatingIp != "" {
		instanceIp.IP = s.FloatingIp
	} else if s.FloatingIpPool != "" {
		// If we have a free floating IP in the pool, use it first
		// rather than creating one
		ui.Say(fmt.Sprintf("Searching for unassociated floating IP in pool %s", s.FloatingIpPool))
		pager := floatingips.List(client)
		err := pager.EachPage(func(page pagination.Page) (bool, error) {
			candidates, err := floatingips.ExtractFloatingIPs(page)

			if err != nil {
				return false, err // stop and throw error out
			}

			for _, candidate := range candidates {
				if candidate.Pool != s.FloatingIpPool || candidate.InstanceID != "" {
					continue // move to next in list
				}

				// In correct pool and able to be allocated
				instanceIp.IP = candidate.IP
				ui.Message(fmt.Sprintf("Selected floating IP: %s", instanceIp.IP))
				state.Put("floatingip_istemp", false)
				return false, nil // stop iterating over pages
			}
			return true, nil // try the next page
		})

		if err != nil {
			err := fmt.Errorf("Error searching for floating ip from pool '%s'", s.FloatingIpPool)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if instanceIp.IP == "" {
			ui.Say(fmt.Sprintf("Creating floating IP..."))
			ui.Message(fmt.Sprintf("Pool: %s", s.FloatingIpPool))
			newIp, err := floatingips.Create(client, floatingips.CreateOpts{
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
			state.Put("floatingip_istemp", true)
		}
	}

	if instanceIp.IP != "" {
		ui.Say(fmt.Sprintf("Associating floating IP with server..."))
		ui.Message(fmt.Sprintf("IP: %s", instanceIp.IP))
		err := floatingips.AssociateInstance(client, server.ID, floatingips.AssociateOpts{
			FloatingIP: instanceIp.IP,
		}).ExtractErr()
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
	instanceIp := state.Get("access_ip").(*floatingips.FloatingIP)

	// Don't delete pool addresses we didn't allocate
	if state.Get("floatingip_istemp") == false {
		return
	}

	// We need the v2 compute client
	client, err := config.computeV2Client()
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error deleting temporary floating IP %s", instanceIp.IP))
		return
	}

	if s.FloatingIpPool != "" && instanceIp.ID != "" {
		if err := floatingips.Delete(client, instanceIp.ID).ExtractErr(); err != nil {
			ui.Error(fmt.Sprintf(
				"Error deleting temporary floating IP %s", instanceIp.IP))
			return
		}

		ui.Say(fmt.Sprintf("Deleted temporary floating IP %s", instanceIp.IP))
	}
}
