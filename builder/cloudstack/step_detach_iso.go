package cloudstack

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

type stepDetachIso struct{}

// Detaches currently ISO file attached to a virtual machine if any.
func (s *stepDetachIso) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	config := state.Get("config").(*Config)

	// Check if state uses iso file and has need to eject it
	if !config.EjectISO || config.SourceISO == "" {
		return multistep.ActionContinue
	}

	ui.Say("Checking attached iso...")

	// Wait to make call detachIso
	if config.EjectISODelay > 0 {
		ui.Message(fmt.Sprintf("Waiting for %v before detaching ISO from virtual machine...", config.EjectISODelay))
		time.Sleep(config.EjectISODelay)
	}

	client := state.Get("client").(*cloudstack.CloudStackClient)

	instanceID, ok := state.Get("instance_id").(string)
	if !ok || instanceID == "" {
		err := fmt.Errorf("Could not retrieve instance_id from state")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message("Detaching iso from virtual machine...")

	// Get a new DetachIsoParams and detaches Iso file from given virtualMachine instance
	detachIsoParams := client.ISO.NewDetachIsoParams(instanceID)
	response, err := client.ISO.DetachIso(detachIsoParams)
	if err != nil || response == nil {
		err := fmt.Errorf("Error detaching ISO from virtual machine: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepDetachIso) Cleanup(state multistep.StateBag) {
	// Nothing to cleanup for this step.
}
