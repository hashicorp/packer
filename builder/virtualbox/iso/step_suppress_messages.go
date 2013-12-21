package iso

import (
	"fmt"
	"github.com/mitchellh/multistep"
	vboxcommon "github.com/mitchellh/packer/builder/virtualbox/common"
	"github.com/mitchellh/packer/packer"
	"log"
)

// This step sets some variables in VirtualBox so that annoying
// pop-up messages don't exist.
type stepSuppressMessages struct{}

func (stepSuppressMessages) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(vboxcommon.Driver)
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

func (stepSuppressMessages) Cleanup(multistep.StateBag) {}
