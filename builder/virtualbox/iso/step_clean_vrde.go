package iso

import (
	"fmt"
	"github.com/mitchellh/multistep"
	vboxcommon "github.com/mitchellh/packer/builder/virtualbox/common"
	"github.com/mitchellh/packer/packer"
)

// Use this step if you want to disable Vrde before generating the artifacts

type stepCleanVrde struct{}

func (s *stepCleanVrde) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	driver := state.Get("driver").(vboxcommon.Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

//	The vboxmanage --vrde flag  takes a string, so we must pass bool
//  to string

	ui.Say(fmt.Sprintf("%t", config.CleanVrde))
	if config.CleanVrde == true {
		var vrdeValue string = "off"
		command := []string{
			"modifyvm", vmName,
			"--vrde", vrdeValue,
		}

		ui.Say(fmt.Sprintf("Setting vrde to %s", vrdeValue))
		err := driver.VBoxManage(command...)
		if err != nil {
			err := fmt.Errorf("Error setting vrde: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *stepCleanVrde) Cleanup(state multistep.StateBag) {}
