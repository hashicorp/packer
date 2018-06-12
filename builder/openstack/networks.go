package openstack

import (
	"fmt"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/attachinterfaces"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/external"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/pagination"
)

// ExternalNetwork is a network with external router.
type ExternalNetwork struct {
	networks.Network
	external.NetworkExternalExt
}

// FindExternalNetwork returns existing network with external router.
// It will return first network if there are many.
func FindExternalNetwork(client *gophercloud.ServiceClient) (*ExternalNetwork, error) {
	var externalNetworks []ExternalNetwork

	allPages, err := networks.List(client, networks.ListOpts{
		Status: "ACTIVE",
	}).AllPages()
	if err != nil {
		return nil, err
	}

	// Extract external networks from found networks.
	err = networks.ExtractNetworksInto(allPages, &externalNetworks)
	if err != nil {
		return nil, err
	}

	if len(externalNetworks) == 0 {
		return nil, fmt.Errorf("no external networks found")
	}

	// Return the first external network.
	return &externalNetworks[0], nil
}

// CheckFloatingIP gets a floating IP by its ID and checks if it is already
// associated with any internal interface.
// It returns floating IP if it can be used.
func CheckFloatingIP(client *gophercloud.ServiceClient, id string) (*floatingips.FloatingIP, error) {
	floatingIP, err := floatingips.Get(client, id).Extract()
	if err != nil {
		return nil, err
	}
	if floatingIP.PortID != "" {
		return nil, fmt.Errorf("provided floating IP '%s' is already associated with port '%s'",
			id, floatingIP.PortID)
	}

	return floatingIP, nil
}

// FindFreeFloatingIP returns free unassociated floating IP.
// It will return first floating IP if there are many.
func FindFreeFloatingIP(client *gophercloud.ServiceClient) (*floatingips.FloatingIP, error) {
	var freeFloatingIP *floatingips.FloatingIP

	pager := floatingips.List(client, floatingips.ListOpts{
		Status: "DOWN",
	})
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		candidates, err := floatingips.ExtractFloatingIPs(page)
		if err != nil {
			return false, err // stop and throw error out
		}

		for _, candidate := range candidates {
			if candidate.PortID != "" {
				continue // this floating IP is associated with port, move to next in list
			}

			// Floating IP is able to be allocated.
			freeFloatingIP = &candidate
			return false, nil // stop iterating over pages
		}
		return true, nil // try the next page
	})
	if err != nil {
		return nil, err
	}
	if freeFloatingIP == nil {
		return nil, fmt.Errorf("no free floating IPs found")
	}

	return freeFloatingIP, nil
}

// GetInstancePortID returns internal port of the instance that can be used for
// the association of a floating IP.
// It will return an ID of a first port if there are many.
func GetInstancePortID(client *gophercloud.ServiceClient, id string) (string, error) {
	interfacesPage, err := attachinterfaces.List(client, id).AllPages()
	if err != nil {
		return "", err
	}
	interfaces, err := attachinterfaces.ExtractInterfaces(interfacesPage)
	if err != nil {
		return "", err
	}
	if len(interfaces) == 0 {
		return "", fmt.Errorf("instance '%s' has no interfaces", id)
	}

	return interfaces[0].PortID, nil
}
