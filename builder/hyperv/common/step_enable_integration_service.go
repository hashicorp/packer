package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepEnableIntegrationService struct {
	name string
}

func (s *StepEnableIntegrationService) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Enabling Integration Service...")

	vmName := state.Get("vmName").(string)
	s.name = "Guest Service Interface"

	err := driver.EnableVirtualMachineIntegrationService(vmName, s.name)

	if err != nil {
		err := fmt.Errorf("Error enabling Integration Service: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepEnableIntegrationService) Cleanup(state multistep.StateBag) {
	// do nothing
}
