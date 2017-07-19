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

type stepCreateFolder struct {
	Folder string
}

func (s *stepCreateFolder) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	finder := state.Get("finder").(*find.Finder)
	dc := state.Get("datacenter").(string)

	if s.Folder != "" {
		ui.Say("Creating or checking destination folders...")

		path := s.Folder
		base := filepath.Join("/", dc, "vm")
		var folders []string
		var root *object.Folder
		var err error
		// We iterate over the path starting with full path
		// If we don't find it, we save the folder name and continue with the previous path
		// The iteration ends when we find an existing path or if we don't find any we'll use
		// the base path
		for {
			root, err = finder.Folder(context.Background(), filepath.ToSlash(filepath.Join(base, path)))
			if err != nil {
				_, folder := filepath.Split(path)
				folders = append(folders, folder)
				if i := strings.LastIndex(path, "/"); i == 0 {
					root, err = finder.Folder(context.Background(), filepath.ToSlash(base))
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
