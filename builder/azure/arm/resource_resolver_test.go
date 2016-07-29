package arm

import (
	"testing"
)

func TestResourceResolverIgnoresEmptyVirtualNetworkName(t *testing.T) {
	c, _, _ := newConfig(getArmBuilderConfiguration(), getPackerConfiguration())
	if c.VirtualNetworkName != "" {
		t.Fatalf("Expected VirtualNetworkName to be empty by default")
	}

	sut := newTestResourceResolver()
	sut.findVirtualNetworkResourceGroup = nil // assert that this is not even called
	sut.Resolve(c)

	if c.VirtualNetworkName != "" {
		t.Fatalf("Expected VirtualNetworkName to be empty")
	}
	if c.VirtualNetworkResourceGroupName != "" {
		t.Fatalf("Expected VirtualNetworkResourceGroupName to be empty")
	}
}

// If the user fully specified the virtual network name and resource group then
// there is no need to do a lookup.
func TestResourceResolverIgnoresSetVirtualNetwork(t *testing.T) {
	c, _, _ := newConfig(getArmBuilderConfiguration(), getPackerConfiguration())
	c.VirtualNetworkName = "--virtual-network-name--"
	c.VirtualNetworkResourceGroupName = "--virtual-network-resource-group-name--"
	c.VirtualNetworkSubnetName = "--virtual-network-subnet-name--"

	sut := newTestResourceResolver()
	sut.findVirtualNetworkResourceGroup = nil // assert that this is not even called
	sut.findVirtualNetworkSubnet = nil        // assert that this is not even called
	sut.Resolve(c)

	if c.VirtualNetworkName != "--virtual-network-name--" {
		t.Fatalf("Expected VirtualNetworkName to be --virtual-network-name--")
	}
	if c.VirtualNetworkResourceGroupName != "--virtual-network-resource-group-name--" {
		t.Fatalf("Expected VirtualNetworkResourceGroupName to be --virtual-network-resource-group-name--")
	}
	if c.VirtualNetworkSubnetName != "--virtual-network-subnet-name--" {
		t.Fatalf("Expected VirtualNetworkSubnetName to be --virtual-network-subnet-name--")
	}
}

// If the user set virtual network name then the code should resolve virtual network
// resource group name.
func TestResourceResolverSetVirtualNetworkResourceGroupName(t *testing.T) {
	c, _, _ := newConfig(getArmBuilderConfiguration(), getPackerConfiguration())
	c.VirtualNetworkName = "--virtual-network-name--"

	sut := newTestResourceResolver()
	sut.Resolve(c)

	if c.VirtualNetworkResourceGroupName != "findVirtualNetworkResourceGroup is mocked" {
		t.Fatalf("Expected VirtualNetworkResourceGroupName to be 'findVirtualNetworkResourceGroup is mocked'")
	}
	if c.VirtualNetworkSubnetName != "findVirtualNetworkSubnet is mocked" {
		t.Fatalf("Expected findVirtualNetworkSubnet to be 'findVirtualNetworkSubnet is mocked'")
	}
}

func newTestResourceResolver() resourceResolver {
	return resourceResolver{
		client: nil,
		findVirtualNetworkResourceGroup: func(*AzureClient, string) (string, error) {
			return "findVirtualNetworkResourceGroup is mocked", nil
		},
		findVirtualNetworkSubnet: func(*AzureClient, string, string) (string, error) {
			return "findVirtualNetworkSubnet is mocked", nil
		},
	}
}
