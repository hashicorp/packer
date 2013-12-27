package iso

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"os"
)

type toolsUploadPathTemplate struct {
	Flavor string
}

type stepUploadTools struct{}

func (*stepUploadTools) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	if config.ToolsUploadFlavor == "" {
		return multistep.ActionContinue
	}

	comm := state.Get("communicator").(packer.Communicator)
	tools_source := state.Get("tools_upload_source").(string)
	ui := state.Get("ui").(packer.Ui)

	ui.Say(fmt.Sprintf("Uploading the '%s' VMware Tools", config.ToolsUploadFlavor))
	f, err := os.Open(tools_source)
	if err != nil {
		state.Put("error", fmt.Errorf("Error opening VMware Tools ISO: %s", err))
		return multistep.ActionHalt
	}
	defer f.Close()

	tplData := &toolsUploadPathTemplate{Flavor: config.ToolsUploadFlavor}
	config.ToolsUploadPath, err = config.tpl.Process(config.ToolsUploadPath, tplData)
	if err != nil {
		err := fmt.Errorf("Error preparing upload path: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err := comm.Upload(config.ToolsUploadPath, f); err != nil {
		state.Put("error", fmt.Errorf("Error uploading VMware Tools: %s", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (*stepUploadTools) Cleanup(multistep.StateBag) {}
