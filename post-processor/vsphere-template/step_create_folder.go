package vsphere_template

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
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
		var root *object.Folder
		var err error

		for {
			root, err = f.Folder(ctx, filepath.ToSlash(filepath.Join(base, path)))
			if err != nil {
				_, folder := filepath.Split(path)
				folders = append(folders, folder)
				if i := strings.LastIndex(path, "/"); i == 0 {
					root, err = f.Folder(ctx, filepath.ToSlash(base))
					if err != nil {
						state.Put("error", err)
						ui.Error(err.Error())
						return multistep.ActionHalt
					}
					break
				} else {
					path = path[:i]
				}
			} else {
				break
			}
		}

		for i := len(folders) - 1; i >= 0; i-- {
			ui.Message(fmt.Sprintf("Creating folder: %v", folders[i]))
			root, err = root.CreateFolder(ctx, folders[i])
			if err != nil {
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}
	return multistep.ActionContinue
}

func (s *StepCreateFolder) Cleanup(multistep.StateBag) {}
