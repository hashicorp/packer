package vmx

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	vmwcommon "github.com/hashicorp/packer/builder/vmware/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/tmp"
)

// StepCloneVMX takes a VMX file and clones the VM into the output directory.
type StepCloneVMX struct {
	OutputDir *string
	Path      string
	VMName    string
	Linked    bool
	tempDir   string
}

func (s *StepCloneVMX) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	halt := func(err error) multistep.StepAction {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	driver := state.Get("driver").(vmwcommon.Driver)
	ui := state.Get("ui").(packersdk.Ui)

	// Set the path we want for the new .vmx file and clone
	vmxPath := filepath.Join(*s.OutputDir, s.VMName+".vmx")
	ui.Say("Cloning source VM...")
	log.Printf("Cloning from: %s", s.Path)
	log.Printf("Cloning to: %s", vmxPath)

	if err := driver.Clone(vmxPath, s.Path, s.Linked); err != nil {
		return halt(err)
	}

	// Read in the machine configuration from the cloned VMX file
	//
	// * The main driver needs the path to the vmx (set above) and the
	// network type so that it can work out things like IP's and MAC
	// addresses
	// * The disk compaction step needs the paths to all attached disks
	if remoteDriver, ok := driver.(vmwcommon.RemoteDriver); ok {
		remoteVmxPath := vmxPath
		tempDir, err := tmp.Dir("packer-vmx")
		if err != nil {
			return halt(err)
		}
		s.tempDir = tempDir
		vmxPath = filepath.Join(tempDir, s.VMName+".vmx")
		if err = remoteDriver.Download(remoteVmxPath, vmxPath); err != nil {
			return halt(err)
		}
	}

	vmxData, err := vmwcommon.ReadVMX(vmxPath)
	if err != nil {
		return halt(err)
	}

	var diskFilenames []string
	// The VMX file stores the path to a configured disk, and information
	// about that disks attachment to a virtual adapter/controller, as a
	// key/value pair.
	// For a virtual disk attached to bus ID 3 of the virtual machines
	// first SCSI adapter the key/value pair would look something like:
	// scsi0:3.fileName = "relative/path/to/scsiDisk.vmdk"
	// The supported adapter types and configuration maximums for each type
	// vary according to the VMware platform type and version, and the
	// Virtual Machine Hardware version used. See the 'Virtual Machine
	// Maximums' section within VMware's 'Configuration Maximums'
	// documentation for each platform:
	// https://kb.vmware.com/s/article/1003497
	// Information about the supported Virtual Machine Hardware versions:
	// https://kb.vmware.com/s/article/1003746
	// The following regexp is used to match all possible disk attachment
	// points that may be found in the VMX file across all VMware
	// platforms/versions and Virtual Machine Hardware versions
	diskPathKeyRe := regexp.MustCompile(`(?i)^(scsi|sata|ide|nvme)[[:digit:]]:[[:digit:]]{1,2}\.fileName`)
	for k, v := range vmxData {
		match := diskPathKeyRe.FindString(k)
		if match != "" && filepath.Ext(v) == ".vmdk" {
			diskFilenames = append(diskFilenames, v)
		}
	}

	// Write out the relative, host filesystem paths to the disks
	var diskFullPaths []string
	for _, diskFilename := range diskFilenames {
		log.Printf("Found attached disk with filename: %s", diskFilename)
		diskFullPaths = append(diskFullPaths, filepath.Join(*s.OutputDir, diskFilename))
	}

	if len(diskFullPaths) == 0 {
		return halt(fmt.Errorf("Could not enumerate disk info from the vmx file"))
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
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}
