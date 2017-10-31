package common

import (
	"fmt"

	"github.com/mitchellh/multistep"
)

type StepUploadTools struct{}

func (c *StepUploadTools) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)

	if err := driver.ToolsInstall(); err != nil {
		state.Put("error", fmt.Errorf("Couldn't mount VMware tools ISO. Please check the 'guest_os_type' in your template.json."))
	}
	return multistep.ActionContinue
}

func (c *StepUploadTools) Cleanup(multistep.StateBag) {}
