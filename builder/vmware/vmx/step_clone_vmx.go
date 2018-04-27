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

type vmxAdapter struct {
	// The string portion of the address used in the vmx file
	strAddr string
	// Max address for adapter, controller, or controller channel
	aAddrMax int
	// Max address for device or channel supported by adapter
	dAddrMax int
}

const (
	// VMware Configuration Maximums - Virtual Hardware Versions 13/14
	//
	// Specifying the max numbers for the adapter/controller:bus/channel
	// *address* as opposed to specifying the maximums as per the VMware
	// documentation allows consistent (inclusive) treatment when looping
	// over each adapter/controller type
	//
	// SCSI - Address range: scsi0:0 to scsi3:15
	scsiAddrName       = "scsi" // String part of address used in the vmx file
	maxSCSIAdapterAddr = 3      // Max 4 adapters
	maxSCSIDeviceAddr  = 15     // Max 15 devices per adapter; ID 7 is the HBA
	// SATA - Address range: sata0:0 to scsi3:29
	sataAddrName       = "sata" // String part of address used in the vmx file
	maxSATAAdapterAddr = 3      // Max 4 controllers
	maxSATADeviceAddr  = 29     // Max 30 devices per controller
	// NVMe - Address range: nvme0:0 to nvme3:14
	nvmeAddrName       = "nvme" // String part of address used in the vmx file
	maxNVMeAdapterAddr = 3      // Max 4 adapters
	maxNVMeDeviceAddr  = 14     // Max 15 devices per adapter
	// IDE - Address range: ide0:0 to ide1:1
	ideAddrName       = "ide" // String part of address used in the vmx file
	maxIDEAdapterAddr = 1     // One controller with primary/secondary channels
	maxIDEDeviceAddr  = 1     // Each channel supports master and slave
)

var (
	scsiAdapter = vmxAdapter{
		strAddr:  scsiAddrName,
		aAddrMax: maxSCSIAdapterAddr,
		dAddrMax: maxSCSIDeviceAddr,
	}
	sataAdapter = vmxAdapter{
		strAddr:  sataAddrName,
		aAddrMax: maxSATAAdapterAddr,
		dAddrMax: maxSATADeviceAddr,
	}
	nvmeAdapter = vmxAdapter{
		strAddr:  nvmeAddrName,
		aAddrMax: maxNVMeAdapterAddr,
		dAddrMax: maxNVMeDeviceAddr,
	}
	ideAdapter = vmxAdapter{
		strAddr:  ideAddrName,
		aAddrMax: maxIDEAdapterAddr,
		dAddrMax: maxIDEDeviceAddr,
	}
)

func (s *StepCloneVMX) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(vmwcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	// Set the path we want for the new .vmx file and clone
	vmxPath := filepath.Join(s.OutputDir, s.VMName+".vmx")
	ui.Say("Cloning source VM...")
	log.Printf("Cloning from: %s", s.Path)
	log.Printf("Cloning to: %s", vmxPath)
	if err := driver.Clone(vmxPath, s.Path); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// Read in the machine configuration from the cloned VMX file
	//
	// * The main driver needs the path to the vmx (set above) and the
	// network type so that it can work out things like IP's and MAC
	// addresses
	// * The disk compaction step needs the paths to all attached disks
	vmxData, err := vmwcommon.ReadVMX(vmxPath)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// Search across all adapter types to get the filenames of attached disks
	allDiskAdapters := []vmxAdapter{
		scsiAdapter,
		sataAdapter,
		nvmeAdapter,
		ideAdapter,
	}
	var diskFilenames []string
	for _, adapter := range allDiskAdapters {
		diskFilenames = append(diskFilenames, getAttachedDisks(adapter, vmxData)...)
	}

	// Write out the relative, host filesystem paths to the disks
	var diskFullPaths []string
	for _, diskFilename := range diskFilenames {
		log.Printf("Found attached disk with filename: %s", diskFilename)
		diskFullPaths = append(diskFullPaths, filepath.Join(s.OutputDir, diskFilename))
	}

	if len(diskFullPaths) == 0 {
		state.Put("error", fmt.Errorf("Could not enumerate disk info from the vmx file"))
		return multistep.ActionHalt
	}

	// Determine the network type by reading out of the .vmx
	var networkType string
	if _, ok := vmxData["ethernet0.connectiontype"]; ok {
		networkType = vmxData["ethernet0.connectiontype"]
		log.Printf("Discovered the network type: %s", networkType)
	}
	if networkType == "" {
		networkType = "nat"
		log.Printf("Defaulting to network type: %s", networkType)
	}

	// Stash all required information in our state bag
	state.Put("vmx_path", vmxPath)
	state.Put("disk_full_paths", diskFullPaths)
	state.Put("vmnetwork", networkType)

	return multistep.ActionContinue
}

func (s *StepCloneVMX) Cleanup(state multistep.StateBag) {
}

func getAttachedDisks(a vmxAdapter, data map[string]string) (attachedDisks []string) {
	// Loop over possible adapter, controller or controller channel
	for x := 0; x <= a.aAddrMax; x++ {
		// Loop over possible addresses for attached devices
		for y := 0; y <= a.dAddrMax; y++ {
			address := fmt.Sprintf("%s%d:%d.filename", a.strAddr, x, y)
			if device, _ := data[address]; filepath.Ext(device) == ".vmdk" {
				attachedDisks = append(attachedDisks, device)
			}
		}
	}
	return
}
