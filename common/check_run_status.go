package common

import (
	"github.com/hashicorp/packer/helper/multistep"
	"errors"
)

func CheckRunStatus(state *multistep.BasicStateBag) error {
	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return rawErr.(error)
	}

	// If we were interrupted or cancelled, then just exit.
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return errors.New("Build was cancelled.")
	}

	if _, ok := state.GetOk(multistep.StateHalted); ok {
		return errors.New("Build was halted.")
	}

	return nil
}
