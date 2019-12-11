package common

import (
	"context"
	"fmt"
	"log"
	"reflect"
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

// Provisioners interpolate most of their fields in the prepare stage; this
// placeholder map helps keep fields that are only generated at build time from
// accidentally being interpolated into empty strings at prepare time.
func PlaceholderData() map[string]string {
	placeholderData := map[string]string{}

	// use reflection to grab the communicator config field off the config
	var sshExample communicator.SSH
	var winrmExample communicator.WinRM

	t := reflect.TypeOf(sshExample)
	n := t.NumField()
	for i := 0; i < n; i++ {
		fVal := t.Field(i)
		name := fVal.Name
		placeholderData[name] = fmt.Sprintf("{{.%s}}", name)
	}

	t = reflect.TypeOf(winrmExample)
	n = t.NumField()
	for i := 0; i < n; i++ {
		fVal := t.Field(i)
		name := fVal.Name
		placeholderData[name] = fmt.Sprintf("{{.%s}}", name)
	}

	placeholderData["ID"] = "{{.ID}}"

	return placeholderData
}

type StepProvision struct {
	Comm packer.Communicator
}

func PopulateProvisionHookData(state multistep.StateBag) map[string]interface{} {
	hookData := map[string]interface{}{}
	// Read communicator data into hook data
	// state.GetOK("id")
	commConf, ok := state.GetOk("communicator_config")
	if !ok {
		log.Printf("Unable to load config from state to populate provisionHookData")
		return hookData
	}
	cast := commConf.(*communicator.Config)

	pd := PlaceholderData()

	v := reflect.ValueOf(cast)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	// Loop over all field values and retrieve them from the ssh config
	for fieldName, _ := range pd {
		fVal := v.FieldByName(fieldName)
		hookData[fieldName] = fVal.Interface()
	}

	return hookData
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
