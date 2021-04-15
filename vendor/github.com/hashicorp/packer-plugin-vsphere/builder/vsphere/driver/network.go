package driver

import (
	"fmt"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type Network struct {
	driver  *VCenterDriver
	network object.NetworkReference
}

func (d *VCenterDriver) NewNetwork(ref *types.ManagedObjectReference) *Network {
	return &Network{
		network: object.NewNetwork(d.client.Client, *ref),
		driver:  d,
	}
}

func (d *VCenterDriver) FindNetwork(name string) (*Network, error) {
	n, err := d.finder.Network(d.ctx, name)
	if err != nil {
		return nil, err
	}
	return &Network{
		network: n,
		driver:  d,
	}, nil
}

func (d *VCenterDriver) FindNetworks(name string) ([]*Network, error) {
	ns, err := d.finder.NetworkList(d.ctx, name)
	if err != nil {
		return nil, err
	}
	var networks []*Network
	for _, n := range ns {
		networks = append(networks, &Network{
			network: n,
			driver:  d,
		})
	}
	return networks, nil
}

func (n *Network) Info(params ...string) (*mo.Network, error) {
	var p []string
	if len(params) == 0 {
		p = []string{"*"}
	} else {
		p = params
	}
	var info mo.Network

	network, ok := n.network.(*object.Network)
	if !ok {
		return nil, fmt.Errorf("unexpected %t network object type", n.network)
	}

	err := network.Properties(n.driver.ctx, network.Reference(), p, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

type MultipleNetworkFoundError struct {
	path   string
	append string
}

func (e *MultipleNetworkFoundError) Error() string {
	return fmt.Sprintf("path '%s' resolves to multiple networks. %s", e.path, e.append)
}
