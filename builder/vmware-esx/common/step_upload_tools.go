package common

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type toolsUploadPathTemplate struct {
	Flavor string
}

type StepUploadTools struct {
	ToolsUploadFlavor string
	ToolsUploadPath   string
	Ctx               interpolate.Context
}

func (c *StepUploadTools) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)

	if c.ToolsUploadFlavor == "" {
		return multistep.ActionContinue
	}

	// Call into the driver to mount the tools
	if err := driver.ToolsInstall(); err != nil {
		state.Put("error", fmt.Errorf("Couldn't mount VMware tools ISO. Please check the 'guest_os_type' in your template.json."))
	}

	// XXX: Although it makes sense to multistep.ActionHalt here, the original step for ESX continues anyways..
	return multistep.ActionContinue
}

func (c *StepUploadTools) Cleanup(multistep.StateBag) {}
