package vmx

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/mitchellh/multistep"
	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
	"github.com/mitchellh/packer/packer"
)

// StepCloneVMX takes a VMX file and clones the VM into the output directory.
type StepCloneVMX struct {
	OutputDir string
	Path      string
	VMName    string
	tempDir   string
}

func (s *StepCloneVMX) Run(state multistep.StateBag) multistep.StepAction {
	halt := func(err error) multistep.StepAction {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	driver := state.Get("driver").(vmwcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	vmxPath := filepath.Join(s.OutputDir, s.VMName+".vmx")

	ui.Say("Cloning source VM...")
	log.Printf("Cloning from: %s", s.Path)
	log.Printf("Cloning to: %s", vmxPath)

	if err := driver.Clone(vmxPath, s.Path); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	if remoteDriver, ok := driver.(vmwcommon.RemoteDriver); ok {
		tempDir, err := ioutil.TempDir("", "packer-vmx")
		if err != nil {
			return halt(err)
		}
		s.tempDir = tempDir
		content, err := remoteDriver.ReadFile(vmxPath)
		if err != nil {
			return halt(err)
		}
		vmxPath = filepath.Join(tempDir, s.VMName+".vmx")
		err = ioutil.WriteFile(vmxPath, content, 0600)
		if err != nil {
			return halt(err)
		}
	}

	vmxData, err := vmwcommon.ReadVMX(vmxPath)
	if err != nil {
		return halt(err)
	}

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
		return halt(fmt.Errorf("Root disk filename could not be found!"))
	}

	state.Put("full_disk_path", filepath.Join(s.OutputDir, diskName))
	state.Put("vmx_path", vmxPath)
	return multistep.ActionContinue
}

func (s *StepCloneVMX) Cleanup(state multistep.StateBag) {
	if s.tempDir != "" {
		os.RemoveAll(s.tempDir)
	}
}
