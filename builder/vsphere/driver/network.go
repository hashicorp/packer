package driver

import (
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type Network struct {
	driver  *Driver
	network *object.Network
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
		network: n.(*object.Network),
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
	err := n.network.Properties(n.driver.ctx, n.network.Reference(), p, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}
