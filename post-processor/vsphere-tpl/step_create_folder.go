package vsphere_tpl

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi/find"
)

type StepCreateFolder struct {
	Folder string
}

func (s *StepCreateFolder) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ctx := state.Get("context").(context.Context)
	f := state.Get("finder").(*find.Finder)
	d := state.Get("datacenter").(string)

	if s.Folder != "" {
		ui.Say("Creating or checking destination folders...")

		if !strings.HasPrefix(s.Folder, "/") {
			s.Folder = filepath.Join("/", s.Folder)
		}

		path := s.Folder
		base := filepath.Join("/", d, "vm")
		var folders []string
		var folder, root string

		for {
			_, err := f.Folder(ctx, filepath.ToSlash(filepath.Join(base, path)))
			if err != nil {

				root, folder = filepath.Split(path)
				folders = append(folders, folder)
				if i := strings.LastIndex(path, "/"); i == 0 {
					break
				} else {
					path = path[:i]
				}
			} else {
				break
			}
		}

		for i := len(folders) - 1; i >= 0; i-- {
			folder, err := f.Folder(ctx, filepath.ToSlash(filepath.Join(base, "/", root)))
			if err != nil {
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}

			ui.Message(fmt.Sprintf("Creating folder: %v", folders[i]))

			if _, err = folder.CreateFolder(ctx, folders[i]); err != nil {
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
			root = filepath.Join(root, folders[i], "/")
		}
	}
	return multistep.ActionContinue
}

func (s *StepCreateFolder) Cleanup(multistep.StateBag) {}
