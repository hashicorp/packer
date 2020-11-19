package common

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

type toolsUploadPathTemplate struct {
	Flavor string
}

type StepUploadTools struct {
	RemoteType        string
	ToolsUploadFlavor string
	ToolsUploadPath   string
	Ctx               interpolate.Context
}

func (c *StepUploadTools) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)

	if c.ToolsUploadFlavor == "" {
		return multistep.ActionContinue
	}

	if c.RemoteType == "esx5" {
		if err := driver.ToolsInstall(); err != nil {
			state.Put("error", fmt.Errorf("Couldn't mount VMware tools ISO. Please check the 'guest_os_type' in your template.json."))
		}
		return multistep.ActionContinue
	}

	comm := state.Get("communicator").(packer.Communicator)
	tools_source := state.Get("tools_upload_source").(string)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say(fmt.Sprintf("Uploading the '%s' VMware Tools", c.ToolsUploadFlavor))
	f, err := os.Open(tools_source)
	if err != nil {
		state.Put("error", fmt.Errorf("Error opening VMware Tools ISO: %s", err))
		return multistep.ActionHalt
	}
	defer f.Close()

	c.Ctx.Data = &toolsUploadPathTemplate{
		Flavor: c.ToolsUploadFlavor,
	}
	c.ToolsUploadPath, err = interpolate.Render(c.ToolsUploadPath, &c.Ctx)
	if err != nil {
		err := fmt.Errorf("Error preparing upload path: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err := comm.Upload(c.ToolsUploadPath, f, nil); err != nil {
		err := fmt.Errorf("Error uploading VMware Tools: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (c *StepUploadTools) Cleanup(multistep.StateBag) {}
