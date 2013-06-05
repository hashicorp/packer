package vmware

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"os/exec"
)

// This step runs the created virtual machine.
//
// Uses:
//   ui     packer.Ui
//   vmx_path string
//
// Produces:
//   <nothing>
type stepRun struct{
	vmxPath string
}

func (s *stepRun) Run(state map[string]interface{}) multistep.StepAction {
	ui := state["ui"].(packer.Ui)
	vmxPath := state["vmx_path"].(string)

	vmrun_path := "/Applications/VMware Fusion.app/Contents/Library/vmrun"

	ui.Say("Starting virtual machine...")
	cmd := exec.Command(vmrun_path, "-T", "fusion", "start", vmxPath, "gui")
	if err := cmd.Run(); err != nil {
		ui.Error(fmt.Sprintf("Error starting VM: %s", err))
		return multistep.ActionHalt
	}

	// Set the VMX path so that we know we started the machine
	s.vmxPath = vmxPath

	return multistep.ActionContinue
}

func (s *stepRun) Cleanup(state map[string]interface{}) {
	ui := state["ui"].(packer.Ui)

	vmrun_path := "/Applications/VMware Fusion.app/Contents/Library/vmrun"

	// If we started the machine... stop it.
	if s.vmxPath != "" {
		ui.Say("Stopping virtual machine...")
		cmd := exec.Command(vmrun_path, "-T", "fusion", "stop", s.vmxPath, "hard")
		if err := cmd.Run(); err != nil {
			ui.Error(fmt.Sprintf("Error stopping VM: %s", err))
		}
	}
}
