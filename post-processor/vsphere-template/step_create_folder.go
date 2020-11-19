package vsphere_template

import (
	"context"
	"fmt"
	"path"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
)

type stepCreateFolder struct {
	Folder string
}

func (s *stepCreateFolder) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	cli := state.Get("client").(*govmomi.Client)
	dcPath := state.Get("dcPath").(string)

	ui.Message("Creating or checking destination folders...")

	base := path.Join(dcPath, "vm")
	fullPath := path.Join(base, s.Folder)
	si := object.NewSearchIndex(cli.Client)

	var folders []string
	var err error
	var ref object.Reference

	// We iterate over the path starting with full path
	// If we don't find it, we save the folder name and continue with the previous path
	// The iteration ends when we find an existing path otherwise it throws error
	for {
		ref, err = si.FindByInventoryPath(context.Background(), fullPath)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		if ref == nil {
			dir, folder := path.Split(fullPath)
			fullPath = path.Clean(dir)

			if fullPath == dcPath {
				err = fmt.Errorf("vSphere base path %s not found", base)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}

			folders = append(folders, folder)
		} else {
			break
		}
	}

	if root, ok := ref.(*object.Folder); ok {
		for i := len(folders) - 1; i >= 0; i-- {
			ui.Message(fmt.Sprintf("Creating folder: %v", folders[i]))

			root, err = root.CreateFolder(context.Background(), folders[i])
			if err != nil {
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}

			fullPath = path.Join(fullPath, folders[i])
		}
		root.SetInventoryPath(fullPath)
		state.Put("folder", root)
	} else {
		err = fmt.Errorf("folder not found: '%v'", ref)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepCreateFolder) Cleanup(multistep.StateBag) {}
