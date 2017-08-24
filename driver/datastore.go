package driver

import (
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/govmomi/vim25/mo"
)

type Datastore struct {
	ds *object.Datastore
	driver *Driver
}

func (d *Driver) NewDatastore(ref *types.ManagedObjectReference) *Datastore {
	return &Datastore{
		ds: object.NewDatastore(d.client.Client, *ref),
		driver: d,
	}
}

func (d *Driver) FindDatastore(name string) (*Datastore, error) {
	ds, err := d.finder.Datastore(d.ctx, name)
	if err != nil {
		return nil, err
	}
	return &Datastore{
		ds:     ds,
		driver: d,
	}, nil
}

func (ds *Datastore) Info(params ...string) (*mo.Datastore, error) {
	var p []string
	if len(params) == 0 {
		p = []string{"*"}
	} else {
		p = params
	}
	var info mo.Datastore
	err := ds.ds.Properties(ds.driver.ctx, ds.ds.Reference(), p, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}
