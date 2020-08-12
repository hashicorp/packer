package driver

import (
	"github.com/vmware/govmomi/vapi/library"
	"github.com/vmware/govmomi/vapi/vcenter"
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

// LibraryTarget specifies a Library or Library item
type LibraryTarget struct {
	LibraryID     string `json:"library_id,omitempty"`
	LibraryItemID string `json:"library_item_id,omitempty"`
}

// CreateSpec info used to create an OVF package from a VM
type CreateSpec struct {
	Description string   `json:"description,omitempty"`
	Name        string   `json:"name,omitempty"`
	Flags       []string `json:"flags,omitempty"`
}

// OVF data used by CreateOVF
type OVF struct {
	Spec   CreateSpec         `json:"create_spec"`
	Source vcenter.ResourceID `json:"source"`
	Target LibraryTarget      `json:"target"`
}

// CreateResult used for decoded a CreateOVF response
type CreateResult struct {
	Succeeded bool                     `json:"succeeded,omitempty"`
	ID        string                   `json:"ovf_library_item_id,omitempty"`
	Error     *vcenter.DeploymentError `json:"error,omitempty"`
}
