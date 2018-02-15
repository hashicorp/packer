package ecs

import (
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

func message(state multistep.StateBag, module string) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	ui := state.Get("ui").(packer.Ui)

	if cancelled || halted {
		ui.Say(fmt.Sprintf("Deleting %s because of cancellation or error...", module))
	} else {
		ui.Say(fmt.Sprintf("Cleaning up '%s'", module))
	}

}
