package virtualbox

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

// This step sets some variables in VirtualBox so that annoying
// pop-up messages don't exist.
type stepSuppressMessages struct{}

func (stepSuppressMessages) Run(state map[string]interface{}) multistep.StepAction {
	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)

	log.Println("Suppressing annoying messages in VirtualBox")
	if err := driver.SuppressMessages(); err != nil {
		err := fmt.Errorf("Error configuring VirtualBox to suppress messages: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (stepSuppressMessages) Cleanup(map[string]interface{}) {}
