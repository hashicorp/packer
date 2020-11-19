package common

import (
	"context"

	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepRemoveFloppy struct {
	Datastore string
	Host      string
}

func (s *StepRemoveFloppy) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(driver.VirtualMachine)
	d := state.Get("driver").(driver.Driver)

	ui.Say("Deleting Floppy drives...")
	floppies, err := vm.FloppyDevices()
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
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
