package iso

import (
	"fmt"

	"github.com/mitchellh/multistep"
	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
	"github.com/mitchellh/packer/packer"
)

type StepRegister struct {
	registeredPath string
}

func (s *StepRegister) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(vmwcommon.Driver)
	ui := state.Get("ui").(packer.Ui)
	vmxPath := state.Get("vmx_path").(string)

	if remoteDriver, ok := driver.(RemoteDriver); ok {
		ui.Say("Registering remote VM...")
		if err := remoteDriver.Register(vmxPath); err != nil {
			err := fmt.Errorf("Error registering VM: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		s.registeredPath = vmxPath
	}

	return multistep.ActionContinue
}

func (s *StepRegister) Cleanup(state multistep.StateBag) {
	if s.registeredPath == "" {
		return
	}

	driver := state.Get("driver").(vmwcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	if remoteDriver, ok := driver.(RemoteDriver); ok {
		ui.Say("Unregistering virtual machine...")
		if err := remoteDriver.Unregister(s.registeredPath); err != nil {
			ui.Error(fmt.Sprintf("Error unregistering VM: %s", err))
		}

		s.registeredPath = ""
	}

}
