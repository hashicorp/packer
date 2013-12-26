package vmx

import (
	"log"
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
}

func (s *StepCloneVMX) Run(state multistep.StateBag) multistep.StepAction {
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

	state.Put("vmx_path", vmxPath)
	return multistep.ActionContinue
}

func (s *StepCloneVMX) Cleanup(state multistep.StateBag) {
}
