package iso

import (
	"fmt"

	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/vmware/govmomi/vim25/types"
	"context"
)

type StepRemoveFloppy struct {
	Datastore string
	Host      string
}

func (s *StepRemoveFloppy) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)
	d := state.Get("driver").(*driver.Driver)

	var UploadedFloppyPath string
	switch s := state.Get("uploaded_floppy_path").(type) {
	case string:
		UploadedFloppyPath = s
	case nil:
		UploadedFloppyPath = ""
	}

	ui.Say("Deleting Floppy drives...")
	devices, err := vm.Devices()
	if err != nil {
		ui.Error(fmt.Sprintf("error removing floppy: %v", err))
		return multistep.ActionHalt
	}
	floppies := devices.SelectByType((*types.VirtualFloppy)(nil))
	if err = vm.RemoveDevice(false, floppies...); err != nil {
		ui.Error(fmt.Sprintf("error removing floppy: %v", err))
		return multistep.ActionHalt
	}

	if UploadedFloppyPath != "" {
		ui.Say("Deleting Floppy image...")
		ds, err := d.FindDatastore(s.Datastore, s.Host)
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		if err := ds.Delete(UploadedFloppyPath); err != nil {
			ui.Error(fmt.Sprintf("Error deleting floppy image '%v': %v", UploadedFloppyPath, err.Error()))
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepRemoveFloppy) Cleanup(state multistep.StateBag) {}
