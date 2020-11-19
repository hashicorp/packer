package ecs

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

func cleanUpMessage(state multistep.StateBag, module string) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	ui := state.Get("ui").(packersdk.Ui)

	if cancelled || halted {
		ui.Say(fmt.Sprintf("Deleting %s because of cancellation or error...", module))
	} else {
		ui.Say(fmt.Sprintf("Cleaning up '%s'", module))
	}
}

func halt(state multistep.StateBag, err error, prefix string) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	if prefix != "" {
		err = fmt.Errorf("%s: %s", prefix, err)
	}

	state.Put("error", err)
	ui.Error(err.Error())
	return multistep.ActionHalt
}

func convertNumber(value int) string {
	if value <= 0 {
		return ""
	}

	return strconv.Itoa(value)
}

func ContainsInArray(arr []string, value string) bool {
	for _, item := range arr {
		if item == value {
			return true
		}
	}

	return false
}
