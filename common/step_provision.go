package common

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
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
type StepProvision struct {
	Comm packer.Communicator
}

type ProvisionHookData struct {
	WinRMPassword string
}

func (s *StepProvision) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	comm := s.Comm
	if comm == nil {
		raw, ok := state.Get("communicator").(packer.Communicator)
		if ok {
			comm = raw.(packer.Communicator)
		}
	}
	hook := state.Get("hook").(packer.Hook)
	ui := state.Get("ui").(packer.Ui)

	// Save data we need to give to provisioners
	WinRMPassword, ok := state.GetOk("winrm_password")
	if !ok {
		WinRMPassword = ""
	} else {
		WinRMPassword = WinRMPassword.(string)
	}
	phd := ProvisionHookData{WinRMPassword.(string)}

	// Run the provisioner in a goroutine so we can continually check
	// for cancellations...
	log.Println("Running the provision hook")
	errCh := make(chan error, 1)
	go func() {
		errCh <- hook.Run(packer.HookProvision, ui, comm, phd)
	}()

	for {
		select {
		case err := <-errCh:
			if err != nil {
				state.Put("error", err)
				return multistep.ActionHalt
			}

			return multistep.ActionContinue
		case <-time.After(1 * time.Second):
			if _, ok := state.GetOk(multistep.StateCancelled); ok {
				log.Println("Cancelling provisioning due to interrupt...")
				hook.Cancel()
				return multistep.ActionHalt
			}
		}
	}
}

func (*StepProvision) Cleanup(multistep.StateBag) {}
