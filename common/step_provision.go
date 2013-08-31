package common

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"time"
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

	// Run the provisioner in a goroutine so we can continually check
	// for cancellations...
	log.Println("Running the provision hook")
	errCh := make(chan error, 1)
	go func() {
		errCh <- hook.Run(packer.HookProvision, ui, comm, nil)
	}()

	for {
		select {
		case err := <-errCh:
			if err != nil {
				state["error"] = err
				return multistep.ActionHalt
			}

			return multistep.ActionContinue
		case <-time.After(1 * time.Second):
			if _, ok := state[multistep.StateCancelled]; ok {
				hook.Cancel()
				return multistep.ActionHalt
			}
		}
	}
}

func (*StepProvision) Cleanup(map[string]interface{}) {}
