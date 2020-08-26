package iso

import (
	"context"

	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/vmware/govmomi/vim25/types"
)

type StepRemoveFloppy struct {
	Datastore string
	Host      string
}

func (s *StepRemoveFloppy) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)
	d := state.Get("driver").(*driver.VCenterDriver)

	ui.Say("Deleting Floppy drives...")
	devices, err := vm.Devices()
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	floppies := devices.SelectByType((*types.VirtualFloppy)(nil))
	if err = vm.RemoveDevice(true, floppies...); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	if UploadedFloppyPath, ok := state.GetOk("uploaded_floppy_path"); ok {
		ui.Say("Deleting Floppy image...")
		ds, err := d.FindDatastore(s.Datastore, s.Host)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
		if err := ds.Delete(UploadedFloppyPath.(string)); err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
		state.Remove("uploaded_floppy_path")
	}

	return multistep.ActionContinue
}

func (s *StepRemoveFloppy) Cleanup(state multistep.StateBag) {}
