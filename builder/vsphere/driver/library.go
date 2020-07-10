package driver

import "github.com/vmware/govmomi/vapi/library"

type Library struct {
	driver  *Driver
	library *library.Library
}

func (d *Driver) FindContentLibrary(name string) (*Library, error) {
	lm := library.NewManager(d.restClient)
	l, err := lm.GetLibraryByName(d.ctx, name)
	if err != nil {
		return nil, err
	}
	return &Library{
		library: l,
		driver:  d,
	}, nil
}
