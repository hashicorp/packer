package common

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

// StepProvision runs the provisioners.
//
// Uses:
//   communicator packer.Communicator
//   hook         packer.Hook
//   ui           packer.Ui
//
// Produces:
//   <nothing>
type StepProvision struct{}

func (*StepProvision) Run(state map[string]interface{}) multistep.StepAction {
	comm := state["communicator"].(packer.Communicator)
	hook := state["hook"].(packer.Hook)
	ui := state["ui"].(packer.Ui)

	log.Println("Running the provision hook")
	if err := hook.Run(packer.HookProvision, ui, comm, nil); err != nil {
		state["error"] = err
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (*StepProvision) Cleanup(map[string]interface{}) {}
