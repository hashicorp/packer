package driver

import (
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"fmt"
)

func (d *Driver) NewResourcePool(ref *types.ManagedObjectReference) *object.ResourcePool {
	return object.NewResourcePool(d.client.Client, *ref)
}

func (d *Driver) ResourcePoolInfo(host *object.ResourcePool, params ...string) (*mo.ResourcePool, error){
	var p []string
	if len(params) == 0 {
		p = []string{"*"}
	} else {
		p = params
	}
	var poolInfo mo.ResourcePool
	err := host.Properties(d.ctx, host.Reference(), p, &poolInfo)
	if err != nil {
		return nil, err
	}
	return &poolInfo, nil
}

func (d *Driver) GetResourcePoolPath(pool *object.ResourcePool) (string, error) {
	f, err := d.ResourcePoolInfo(pool, "name", "parent")
	if err != nil {
		return "", err
	}
	if f.Parent.Type == "ComputeResource" {
		return "", nil
	} else {
		parent := d.NewResourcePool(f.Parent)
		parentPath, err := d.GetResourcePoolPath(parent)
		if err != nil {
			return "", err
		}
		if parentPath == "" {
			return f.Name, nil
		} else {
			return fmt.Sprintf("%v/%v", parentPath, f.Name), nil
		}
	}
}
