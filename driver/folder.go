package driver

import (
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

func (d *Driver) NewFolder(ref *types.ManagedObjectReference) *object.Folder {
	return object.NewFolder(d.client.Client, *ref)
}

func (d *Driver) GetFolderPath(folder *object.Folder) (string, error) {
	return folder.ObjectName(d.ctx)
}
