package vmx

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// StepCloneVMX takes a VMX file and clones the VM into the output directory.
type StepCloneVMX struct {
	OutputDir string
	Path      string
	VMName    string
}

func (s *StepCloneVMX) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	vmxPath := filepath.Join(s.OutputDir, s.VMName+".vmx")

	ui.Say("Cloning VMX...")
	log.Printf("Cloning from: %s", s.Path)
	log.Printf("Cloning to: %s", vmxPath)

	from, err := os.Open(s.Path)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	defer from.Close()

	to, err := os.Create(vmxPath)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	defer to.Close()

	if _, err := io.Copy(to, from); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	state.Put("vmx_path", vmxPath)
	return multistep.ActionContinue
}

func (s *StepCloneVMX) Cleanup(state multistep.StateBag) {
}
