package iso

import (
	"fmt"

	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi/vim25/types"
)

type StepRemoveFloppy struct {
	Datastore          string
	Host               string
	UploadedFloppyPath string
}

func (s *StepRemoveFloppy) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)
	d := state.Get("driver").(*driver.Driver)

	devices, err := vm.Devices()
	if err != nil {
		ui.Error(fmt.Sprintf("error removing floppy: %v", err))
		return multistep.ActionHalt
	}
	cdroms := devices.SelectByType((*types.VirtualFloppy)(nil))
	if err = vm.RemoveDevice(false, cdroms...); err != nil {
		ui.Error(fmt.Sprintf("error removing floppy: %v", err))
		return multistep.ActionHalt
	}

	if s.UploadedFloppyPath != "" {
		ds, err := d.FindDatastore(s.Datastore, s.Host)
		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		if err := ds.Delete(s.UploadedFloppyPath); err != nil {
			ui.Error(fmt.Sprintf("Error deleting floppy image '%v': %v", s.UploadedFloppyPath, err.Error()))
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepRemoveFloppy) Cleanup(state multistep.StateBag) {}
