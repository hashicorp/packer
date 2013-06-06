package vmware

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step waits for SSH to become available and establishes an SSH
// connection.
//
// Uses:
//   config *config
//   ui     packer.Ui
//
// Produces:
//   <nothing>
type stepWaitForSSH struct{}

func (stepWaitForSSH) Run(state map[string]interface{}) multistep.StepAction {
	ui := state["ui"].(packer.Ui)

	ui.Say("Waiting for SSH to become available...")

	for {
		// First we wait for the IP to become available...
	}

	return multistep.ActionContinue
}

func (stepWaitForSSH) Cleanup(map[string]interface{}) {}
