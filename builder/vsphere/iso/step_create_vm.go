package iso

import (
	"fmt"

	vspcommon "github.com/hashicorp/packer/builder/vsphere/common"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

// This step configures a VM by setting some default settings as well
// as taking in custom data to set, attaching a floppy if it exists, etc.
//
// Uses:
//   vmx_path string
type stepCreateVM struct {
	VMName             string
	Folder             string
	Datastore          string
	Cpu                uint
	MemSize            uint
	DiskSize           uint
	AdditionalDiskSize []uint
	DiskThick          bool
	GuestType          string
	NetworkName        string
	NetworkAdapter     string
	Annotation         string
}

func (s *stepCreateVM) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(vspcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating VM")
	if err := driver.CreateVirtualMachine(
		s.VMName,
		s.Folder,
		s.Datastore,
		s.Cpu,
		s.MemSize,
		s.DiskSize,
		s.DiskThick,
		s.GuestType,
		s.NetworkName,
		s.NetworkAdapter,
		s.Annotation); err != nil {
		state.Put("error", err)
		ui.Error(fmt.Sprintf("Error creating VM: %s", err))
		return multistep.ActionHalt
	}

	//Set a floppy disk, but only if we have one
	if floppyPath, ok := state.GetOk("floppy_path"); ok {
		ui.Say("Inserting floppy")

		floppyDevice, err := driver.AddFloppy(floppyPath.(string))
		if err != nil {
			state.Put("error", err)
			ui.Error(fmt.Sprintf("Error inserting floppy: %s", err))
			return multistep.ActionHalt
		}
		state.Put("floppy_device", floppyDevice)
	}

	ui.Say("Mounting ISO")
	isoPath := state.Get("iso_path")
	cdromDevice, err := driver.MountISO(isoPath.(string))
	if err != nil {
		state.Put("error", err)
		ui.Error(fmt.Sprintf("Error mounting ISO: %s", err))
		return multistep.ActionHalt
	}
	state.Put("cdrom_device", cdromDevice)

	if len(s.AdditionalDiskSize) > 0 {
		ui.Say("Creating additional hard drives...")
		for _, additionalsize := range s.AdditionalDiskSize {
			size := additionalsize
			if err := driver.CreateDisk(size, s.DiskThick); err != nil {
				err := fmt.Errorf("Error creating additional disk: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	return multistep.ActionContinue
}

func (s *stepCreateVM) Cleanup(state multistep.StateBag) {
}
