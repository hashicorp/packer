package vmware

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step creates the virtual disks for the VM.
//
// Uses:
//   config *config
//   ui     packer.Ui
//
// Produces:
//   <nothing>
type stepWaitForIP struct{}

func (stepWaitForIP) Run(state map[string]interface{}) multistep.StepAction {
	ui := state["ui"].(packer.Ui)

	ui.Say("Waiting for SSH to become available...")
	select{}

	return multistep.ActionContinue
}

func (stepWaitForIP) Cleanup(map[string]interface{}) {}
