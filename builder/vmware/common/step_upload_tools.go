package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"os"
)

type toolsUploadPathTemplate struct {
	Flavor string
}

type StepUploadTools struct {
	RemoteType        string
	ToolsUploadFlavor string
	ToolsUploadPath   string
	Tpl               *packer.ConfigTemplate
}

func (c *StepUploadTools) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)

	if c.RemoteType == "esx5" {
		if err := driver.ToolsInstall(); err != nil {
			state.Put("error", fmt.Errorf("Couldn't mount VMware tools ISO."))
		}
		return multistep.ActionContinue
	}

	if c.ToolsUploadFlavor == "" {
		return multistep.ActionContinue
	}

	comm := state.Get("communicator").(packer.Communicator)
	tools_source := state.Get("tools_upload_source").(string)
	ui := state.Get("ui").(packer.Ui)

	ui.Say(fmt.Sprintf("Uploading the '%s' VMware Tools", c.ToolsUploadFlavor))
	f, err := os.Open(tools_source)
	if err != nil {
		state.Put("error", fmt.Errorf("Error opening VMware Tools ISO: %s", err))
		return multistep.ActionHalt
	}
	defer f.Close()

	tplData := &toolsUploadPathTemplate{
		Flavor: c.ToolsUploadFlavor,
	}
	c.ToolsUploadPath, err = c.Tpl.Process(c.ToolsUploadPath, tplData)
	if err != nil {
		err := fmt.Errorf("Error preparing upload path: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err := comm.Upload(c.ToolsUploadPath, f); err != nil {
		err := fmt.Errorf("Error uploading VMware Tools: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (c *StepUploadTools) Cleanup(multistep.StateBag) {}
