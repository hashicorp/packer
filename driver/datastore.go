package driver

import (
	"github.com/vmware/govmomi/object"
)

type Datastore struct {
	ds *object.Datastore
	driver *Driver
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
