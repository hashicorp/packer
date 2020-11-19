package common

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// This step suppresses any messages that VMware product might show.
type StepSuppressMessages struct{}

func (s *StepSuppressMessages) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmxPath := state.Get("vmx_path").(string)

	log.Println("Suppressing messages in VMX")
	if err := driver.SuppressMessages(vmxPath); err != nil {
		err := fmt.Errorf("Error suppressing messages: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepSuppressMessages) Cleanup(state multistep.StateBag) {}
