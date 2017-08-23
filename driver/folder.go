package driver

import (
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/govmomi/vim25/mo"
	"fmt"
)

func (d *Driver) NewFolder(ref *types.ManagedObjectReference) *object.Folder {
	return object.NewFolder(d.client.Client, *ref)
}

func (d *Driver) FolderInfo(folder *object.Folder, params ...string) (*mo.Folder, error) {
	var p []string
	if len(params) == 0 {
		p = []string{"*"}
	} else {
		p = params
	}
	var folderInfo mo.Folder
	err := folder.Properties(d.ctx, folder.Reference(), p, &folderInfo)
	if err != nil {
		return nil, err
	}
	return &folderInfo, nil
}

func (d *Driver) GetFolderPath(folder *object.Folder) (string, error) {
	f, err := d.FolderInfo(folder, "name", "parent")
	if err != nil {
		return "", err
	}
	if f.Parent.Type == "Datacenter" {
		return "", nil
	} else {
		parent := d.NewFolder(f.Parent)
		parentPath, err := d.GetFolderPath(parent)
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
