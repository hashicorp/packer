package common

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepSetBootOrder struct {
	BootOrder []string
}

func (s *StepSetBootOrder) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmName := state.Get("vmName").(string)

	if s.BootOrder != nil {
		ui.Say(fmt.Sprintf("Setting boot order to %q", s.BootOrder))
		err := driver.SetBootOrder(vmName, s.BootOrder)

		if err != nil {
			err := fmt.Errorf("Error setting the boot order: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepSetBootOrder) Cleanup(state multistep.StateBag) {
	// do nothing
}
