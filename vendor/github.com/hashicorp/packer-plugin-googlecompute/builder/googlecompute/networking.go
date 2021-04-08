package googlecompute

import (
	"fmt"
	"strings"
)

// This method will build a network and subnetwork ID from the provided
// instance config, and return them in that order.
func getNetworking(c *InstanceConfig) (string, string, error) {
	networkId := c.Network
	subnetworkId := c.Subnetwork

	// Apply network naming requirements per
	// https://cloud.google.com/compute/docs/reference/latest/instances#resource
	switch c.Network {
	// It is possible to omit the network property as long as a subnet is
	// specified. That will be validated later.
	case "":
		break
	// This special short name should be expanded.
	case "default":
		networkId = "global/networks/default"
	// A value other than "default" was provided for the network name.
	default:
		// If the value doesn't contain a slash, we assume it's not a full or
		// partial URL. We will expand it into a partial URL here and avoid
		// making an API call to discover the network as it's common for the
		// caller to not have permission against network discovery APIs.
		if !strings.Contains(c.Network, "/") {
			networkId = "projects/" + c.NetworkProjectId + "/global/networks/" + c.Network
		}
	}

	// Apply subnetwork naming requirements per
	// https://cloud.google.com/compute/docs/reference/latest/instances#resource
	switch c.Subnetwork {
	case "":
		// You can't omit both subnetwork and network
		if networkId == "" {
			return networkId, subnetworkId, fmt.Errorf("both network and subnetwork were empty.")
		}
		// An empty subnetwork is only valid for networks in legacy mode or
		// auto-subnet mode. We could make an API call to get that information
		// about the network, but it's common for the caller to not have
		// permission to that API. We'll proceed assuming they're correct in
		// omitting the subnetwork and let the compute.insert API surface an
		// error about an invalid network configuration if it exists.
		break
	default:
		// If the value doesn't contain a slash, we assume it's not a full or
		// partial URL. We will expand it into a partial URL here and avoid
		// making a call to discover the subnetwork.
		if !strings.Contains(c.Subnetwork, "/") {
			subnetworkId = "projects/" + c.NetworkProjectId + "/regions/" + c.Region + "/subnetworks/" + c.Subnetwork
		}
	}
	return networkId, subnetworkId, nil
}
