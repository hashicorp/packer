package common

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
)

// This step sets some variables in VirtualBox so that annoying
// pop-up messages don't exist.
type StepSuppressMessages struct{}

func (StepSuppressMessages) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	log.Println("Suppressing annoying messages in VirtualBox")
	if err := driver.SuppressMessages(); err != nil {
		err := fmt.Errorf("Error configuring VirtualBox to suppress messages: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (StepSuppressMessages) Cleanup(multistep.StateBag) {}
