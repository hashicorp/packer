package driver

import (
	"fmt"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type Network struct {
	driver  *Driver
	network object.NetworkReference
}

func (d *Driver) NewNetwork(ref *types.ManagedObjectReference) *Network {
	return &Network{
		network: object.NewNetwork(d.client.Client, *ref),
		driver:  d,
	}
}

func (d *Driver) FindNetwork(name string) (*Network, error) {
	n, err := d.finder.Network(d.ctx, name)
	if err != nil {
		return nil, err
	}
	return &Network{
		network: n,
		driver:  d,
	}, nil
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
