package ecs

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

func message(state multistep.StateBag, module string) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	ui := state.Get("ui").(packer.Ui)

	if cancelled || halted {
		ui.Say(fmt.Sprintf("Delete the %s because cancelation or error...", module))
	} else {
		ui.Say(fmt.Sprintf("Clean the created %s", module))
	}

}
