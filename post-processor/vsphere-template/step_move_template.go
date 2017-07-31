package vsphere_template

import (
	"context"
	"path"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

type stepMoveTemplate struct {
	Folder string
}

func (s *stepMoveTemplate) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	finder := state.Get("finder").(*find.Finder)
	dcPath := state.Get("dcPath").(string)
	vm := state.Get("vm").(*object.VirtualMachine)

	if s.Folder != "" {
		ui.Say("Moving template...")

		folder, err := finder.Folder(context.Background(), path.Join(dcPath, "vm", s.Folder))
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())

			return multistep.ActionHalt
		}

		task, err := folder.MoveInto(context.Background(), []types.ManagedObjectReference{vm.Reference()})
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		if err = task.Wait(context.Background()); err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}
	return multistep.ActionContinue
}

func (s *stepMoveTemplate) Cleanup(multistep.StateBag) {}
