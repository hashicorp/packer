package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

// This step suppresses any messages that VMware product might show.
type StepSuppressMessages struct{}

func (s *StepSuppressMessages) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
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
