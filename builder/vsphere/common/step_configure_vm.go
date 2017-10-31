package common

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

// This step configures a VM by setting some default settings as well
// as taking in custom data to set, attaching a floppy if it exists, etc.
//
// Uses:
//   vmx_path string
type StepConfigureVM struct {
	CustomData map[string]string
}

func (s *StepConfigureVM) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	// Set this so that no dialogs ever appear from Packer.
	if err := driver.VMChange("msg.autoanswer=true"); err != nil {
		state.Put("error", err)
		ui.Error(fmt.Sprintf("Error changing VM: %s", err))
		return multistep.ActionHalt
	}
	// Set custom data
	for k, v := range s.CustomData {
		log.Printf("Setting VMX: '%s' = '%s'", k, v)
		k = strings.ToLower(k)
		if err := driver.VMChange(fmt.Sprintf("%s=%s", k, v)); err != nil {
			state.Put("error", err)
			ui.Error(fmt.Sprintf("Error changing VM: %s", err))
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepConfigureVM) Cleanup(state multistep.StateBag) {
}
