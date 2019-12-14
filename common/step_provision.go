package common

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hashicorp/packer/helper/communicator"
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

func PopulateProvisionHookData(state multistep.StateBag) map[string]interface{} {
	// Load Builder hook data from state, if it has been set.
	// hookData := state.GetOk("generated_data").(map[string]interface{})

	// instance_id is placed in state by the builders.
	// Not yet implemented in Chroot, lxc/lxd, Azure, Qemu.
	// Implemented in most others including digitalOcean (droplet id),
	// docker (container_id), and clouds which use "server" internally instead
	// of instance.

	// Also note that Chroot and lxc/lxd builders tend to have their own custom
	// step_provision, so they won't use this code path.
	id, ok := state.GetOk("instance_id")
	if ok {
		hookData["ID"] = id
	} else {
		// Warn user that the id isn't implemented
		hookData["ID"] = "ERR_ID_NOT_IMPLEMENTED_BY_BUILDER"
	}
	hookData["PACKER_RUN_UUID"] = os.Getenv("PACKER_RUN_UUID")

	// Read communicator data into hook data
	comm, ok := state.GetOk("communicator_config")
	if !ok {
		log.Printf("Unable to load config from state to populate provisionHookData")
		return hookData
	}
	commConf := comm.(*communicator.Config)

	// Loop over all field values and retrieve them from the ssh config
	hookData["Host"] = commConf.Host()
	hookData["Port"] = commConf.Port()
	hookData["User"] = commConf.User()
	hookData["Password"] = commConf.Password()
	hookData["ConnType"] = commConf.Type
	hookData["SSHPublicKey"] = commConf.SSHPublicKey
	hookData["SSHPrivateKey"] = commConf.SSHPrivateKey

	// Backwards compatibility; in practice, WinRMPassword is fulfilled by
	// Password.
	hookData["WinRMPassword"] = commConf.WinRMPassword

	return hookData
}

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

	hookData := PopulateProvisionHookData(state)

	// Run the provisioner in a goroutine so we can continually check
	// for cancellations...
	if hooktype == packer.HookProvision {
		log.Println("Running the provision hook")
	} else if hooktype == packer.HookCleanupProvision {
		ui.Say("Provisioning step had errors: Running the cleanup provisioner, if present...")
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- hook.Run(ctx, hooktype, ui, comm, hookData)
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
