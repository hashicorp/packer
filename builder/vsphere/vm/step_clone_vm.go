package vm

import (
	"fmt"
	"log"

	vspcommon "github.com/hashicorp/packer/builder/vsphere/common"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type stepCloneVM struct {
	SrcVMName      string
	VMName         string
	Folder         string
	Datastore      string
	Cpu            uint
	MemSize        uint
	DiskSize       uint
	DiskThick      bool
	NetworkName    string
	NetworkAdapter string
	Annotation     string
}

func (s *stepCloneVM) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(vspcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Cloning source VM...")
	log.Printf("Cloning from: %s", s.SrcVMName)
	log.Printf("Cloning to: %s", s.VMName)
	if err := driver.CloneVirtualMachine(
		s.SrcVMName,
		s.VMName,
		s.Folder,
		s.Datastore,
		s.Cpu,
		s.MemSize,
		s.DiskSize,
		s.DiskThick,
		s.NetworkName,
		s.NetworkAdapter,
		s.Annotation); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	if len(config.AdditionalDiskSize) > 0 {
		ui.Say("Creating additional hard drives...")
		for _, additionalsize := range config.AdditionalDiskSize {
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

func (s *stepCloneVM) Cleanup(state multistep.StateBag) {
}
