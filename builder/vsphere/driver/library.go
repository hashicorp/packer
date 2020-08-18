package driver

import (
	"github.com/vmware/govmomi/vapi/library"
)

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

func (d *Driver) FindContentLibraryItem(libraryId string, name string) (*library.Item, error) {
	lm := library.NewManager(d.restClient)
	items, err := lm.GetLibraryItems(d.ctx, libraryId)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.Name == name {
			return &item, nil
		}
	}
	return nil, nil
}
