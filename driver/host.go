package driver

import (
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func (d *Driver) NewHost(ref *types.ManagedObjectReference) *object.HostSystem {
	return object.NewHostSystem(d.client.Client, *ref)
}

func (d *Driver) HostInfo(host *object.HostSystem, params ...string) (*mo.HostSystem, error){
	var p []string
	if len(params) == 0 {
		p = []string{"*"}
	} else {
		p = params
	}
	var hostInfo mo.HostSystem
	err := host.Properties(d.ctx, host.Reference(), p, &hostInfo)
	if err != nil {
		return nil, err
	}
	return &hostInfo, nil
}
