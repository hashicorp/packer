package vsphere_template

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
)

type stepCreateFolder struct {
	Folder string
}

func (s *stepCreateFolder) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	cli := state.Get("client").(*govmomi.Client)
	dcPath := state.Get("dcPath").(string)

	if s.Folder != "" {
		ui.Say("Creating or checking destination folders...")

		base := filepath.Join(dcPath, "vm")
		path := filepath.ToSlash(filepath.Join(base, s.Folder))
		si := object.NewSearchIndex(cli.Client)

		var folders []string
		var err error
		var ref object.Reference

		// We iterate over the path starting with full path
		// If we don't find it, we save the folder name and continue with the previous path
		// The iteration ends when we find an existing path otherwise it throws error
		for {
			ref, err = si.FindByInventoryPath(context.Background(), path)
			if err != nil {
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
			if ref == nil {
				_, folder := filepath.Split(path)
				folders = append(folders, folder)
				path = path[:strings.LastIndex(path, "/")]

				if path == dcPath {
					err = fmt.Errorf("vSphere base path %s not found", filepath.ToSlash(base))
					state.Put("error", err)
					ui.Error(err.Error())
					return multistep.ActionHalt
				}
			} else {
				break
			}
		}

		root := ref.(*object.Folder)
		for i := len(folders) - 1; i >= 0; i-- {
			ui.Message(fmt.Sprintf("Creating folder: %v", folders[i]))
			root, err = root.CreateFolder(context.Background(), folders[i])
			if err != nil {
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}
	return multistep.ActionContinue
}

func (s *stepCreateFolder) Cleanup(multistep.StateBag) {}
