package vmx

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	vmwcommon "github.com/hashicorp/packer/builder/vmware/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepCloneVMX takes a VMX file and clones the VM into the output directory.
type StepCloneVMX struct {
	OutputDir string
	Path      string
	VMName    string
}

func (s *StepCloneVMX) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(vmwcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	// initially we need to stash the path to the original .vmx file
	vmxPath := filepath.Join(s.OutputDir, s.VMName+".vmx")

	// so first, let's clone the source path to the vmxPath
	ui.Say("Cloning source VM...")
	log.Printf("Cloning from: %s", s.Path)
	log.Printf("Cloning to: %s", vmxPath)
	if err := driver.Clone(vmxPath, s.Path); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	ui.Say("Successfully cloned source VM to: %s", vmxPath)

	// now we read the .vmx so we can determine what else to stash
	vmxData, err := vmwcommon.ReadVMX(vmxPath)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// figure out the disk filename by walking through all device types
	var diskName string
	if _, ok := vmxData["scsi0:0.filename"]; ok {
		diskName = vmxData["scsi0:0.filename"]
	}
	if _, ok := vmxData["sata0:0.filename"]; ok {
		diskName = vmxData["sata0:0.filename"]
	}
	if _, ok := vmxData["ide0:0.filename"]; ok {
		diskName = vmxData["ide0:0.filename"]
	}
	if diskName == "" {
		err := fmt.Errorf("Root disk filename could not be found!")
		state.Put("error", err)
		return multistep.ActionHalt
	}
	log.Printf("Found root disk filename: %s", diskName)

	// determine the network type by reading out of the .vmx
	var networkType string
	if _, ok := vmxData["ethernet0.connectiontype"]; ok {
		networkType = vmxData["ethernet0.connectiontype"]
		log.Printf("Discovered the network type: %s", networkType)
	}
	if networkType == "" {
		networkType = "nat"
		log.Printf("Defaulting to network type: %s", networkType)
	}
	ui.Say("Using network type: %s", networkType)

	// we were able to find everything, so stash it in our state.
	state.Put("vmx_path", vmxPath)
	state.Put("full_disk_path", filepath.Join(s.OutputDir, diskName))
	state.Put("vmnetwork", networkType)

	return multistep.ActionContinue
}

func (s *StepCloneVMX) Cleanup(state multistep.StateBag) {
}
