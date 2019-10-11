package common

import (
	"context"
	"fmt"
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

func (s *StepProvision) runWithHook(ctx context.Context, state multistep.StateBag, hooktype string) multistep.StepAction {
	// hooktype will be either packer.HookProvision or packer.HookCleanupProvision
	comm := s.Comm
	if comm == nil {
		raw, ok := state.Get("communicator").(packer.Communicator)
		if ok {
			comm = raw.(packer.Communicator)
		}
	}
	hook := state.Get("hook").(packer.Hook)
	ui := state.Get("ui").(packer.Ui)

	// Run the provisioner in a goroutine so we can continually check
	// for cancellations...
	if hooktype == packer.HookProvision {
		log.Println("Running the provision hook")
	} else if hooktype == packer.HookCleanupProvision {
		ui.Say("Provisioning step had errors: Running the cleanup provisioner, if present...")
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- hook.Run(ctx, hooktype, ui, comm, nil)
	}()

	for {
		select {
		case err := <-errCh:
			if err != nil {
				if hooktype == packer.HookProvision {
					// We don't overwrite the error if it's a cleanup
					// provisioner being run.
					state.Put("error", err)
				} else if hooktype == packer.HookCleanupProvision {
					origErr := state.Get("error").(error)
					state.Put("error", fmt.Errorf("Cleanup failed: %s. "+
						"Original Provisioning error: %s", err, origErr))
				}
				return multistep.ActionHalt
			}

			return multistep.ActionContinue
		case <-ctx.Done():
			log.Printf("Cancelling provisioning due to context cancellation: %s", ctx.Err())
			return multistep.ActionHalt
		case <-time.After(1 * time.Second):
			if _, ok := state.GetOk(multistep.StateCancelled); ok {
				log.Println("Cancelling provisioning due to interrupt...")
				return multistep.ActionHalt
			}
		}
	}
}

func (s *StepProvision) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	return s.runWithHook(ctx, state, packer.HookProvision)
}

func (s *StepProvision) Cleanup(state multistep.StateBag) {
	// We have a "final" provisioner that gets defined by "error-cleanup-provisioner"
	// which we only call if there's an error during the provision run and
	// the "error-cleanup-provisioner" is defined.
	if _, ok := state.GetOk("error"); ok {
		s.runWithHook(context.Background(), state, packer.HookCleanupProvision)
	}
}
