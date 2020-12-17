package commonsteps

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepProvision runs the provisioners.
//
// Uses:
//   communicator packersdk.Communicator
//   hook         packersdk.Hook
//   ui           packersdk.Ui
//
// Produces:
//   <nothing>

const HttpIPNotImplemented = "ERR_HTTP_IP_NOT_IMPLEMENTED_BY_BUILDER"
const HttpPortNotImplemented = "ERR_HTTP_PORT_NOT_IMPLEMENTED_BY_BUILDER"
const HttpAddrNotImplemented = "ERR_HTTP_ADDR_NOT_IMPLEMENTED_BY_BUILDER"

func PopulateProvisionHookData(state multistep.StateBag) map[string]interface{} {
	hookData := make(map[string]interface{})

	// Load Builder hook data from state, if it has been set.
	hd, ok := state.GetOk("generated_data")
	if ok {
		hookData = hd.(map[string]interface{})
	}

	// Warn user that the id isn't implemented
	hookData["ID"] = "ERR_ID_NOT_IMPLEMENTED_BY_BUILDER"

	// instance_id is placed in state by the builders.
	// Not yet implemented in Chroot, lxc/lxd, Azure, Qemu.
	// Implemented in most others including digitalOcean (droplet id),
	// docker (container_id), and clouds which use "server" internally instead
	// of instance.
	id, ok := state.GetOk("instance_id")
	if ok {
		hookData["ID"] = id
	}

	hookData["PackerRunUUID"] = os.Getenv("PACKER_RUN_UUID")

	// Packer HTTP info
	hookData["PackerHTTPIP"] = HttpIPNotImplemented
	hookData["PackerHTTPPort"] = HttpPortNotImplemented
	hookData["PackerHTTPAddr"] = HttpAddrNotImplemented

	httpPort, okPort := state.GetOk("http_port")
	if okPort {
		hookData["PackerHTTPPort"] = strconv.Itoa(httpPort.(int))
	}
	httIP, okIP := state.GetOk("http_ip")
	if okIP {
		hookData["PackerHTTPIP"] = httIP.(string)
	}
	if okPort && okIP {
		hookData["PackerHTTPAddr"] = fmt.Sprintf("%s:%s", hookData["PackerHTTPIP"], hookData["PackerHTTPPort"])
	}

	// Read communicator data into hook data
	comm, ok := state.GetOk("communicator_config")
	if !ok {
		log.Printf("Unable to load communicator config from state to populate provisionHookData")
		return hookData
	}
	commConf := comm.(*communicator.Config)

	// Loop over all field values and retrieve them from the ssh config
	hookData["Host"] = commConf.Host()
	hookData["Port"] = commConf.Port()
	hookData["User"] = commConf.User()
	hookData["Password"] = commConf.Password()
	hookData["ConnType"] = commConf.Type
	hookData["SSHPublicKey"] = string(commConf.SSHPublicKey)
	hookData["SSHPrivateKey"] = string(commConf.SSHPrivateKey)
	hookData["SSHPrivateKeyFile"] = commConf.SSHPrivateKeyFile
	hookData["SSHAgentAuth"] = commConf.SSHAgentAuth

	// Backwards compatibility; in practice, WinRMPassword is fulfilled by
	// Password.
	hookData["WinRMPassword"] = commConf.WinRMPassword

	return hookData
}

type StepProvision struct {
	Comm packersdk.Communicator
}

func (s *StepProvision) runWithHook(ctx context.Context, state multistep.StateBag, hooktype string) multistep.StepAction {
	// hooktype will be either packersdk.HookProvision or packersdk.HookCleanupProvision
	comm := s.Comm
	if comm == nil {
		raw, ok := state.Get("communicator").(packersdk.Communicator)
		if ok {
			comm = raw.(packersdk.Communicator)
		}
	}

	hook := state.Get("hook").(packersdk.Hook)
	ui := state.Get("ui").(packersdk.Ui)

	hookData := PopulateProvisionHookData(state)

	// Update state generated_data with complete hookData
	// to make them accessible by post-processors
	state.Put("generated_data", hookData)

	// Run the provisioner in a goroutine so we can continually check
	// for cancellations...
	if hooktype == packersdk.HookProvision {
		log.Println("Running the provision hook")
	} else if hooktype == packersdk.HookCleanupProvision {
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
				if hooktype == packersdk.HookProvision {
					// We don't overwrite the error if it's a cleanup
					// provisioner being run.
					state.Put("error", err)
				} else if hooktype == packersdk.HookCleanupProvision {
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
	return s.runWithHook(ctx, state, packersdk.HookProvision)
}

func (s *StepProvision) Cleanup(state multistep.StateBag) {
	// We have a "final" provisioner that gets defined by "error-cleanup-provisioner"
	// which we only call if there's an error during the provision run and
	// the "error-cleanup-provisioner" is defined.
	if _, ok := state.GetOk("error"); ok {
		s.runWithHook(context.Background(), state, packersdk.HookCleanupProvision)
	}
}
