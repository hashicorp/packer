package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"os"
)

type StepPrepareTools struct{
	RemoteType        string `mapstructure:"remote_type"`
	ToolsUploadFlavor string   `mapstructure:"tools_upload_flavor"`
	ToolsUploadPath   string   `mapstructure:"tools_upload_path"`
}

func (c *StepPrepareTools) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)

	if c.RemoteType == "esx5" {
		return multistep.ActionContinue
	}

	if c.ToolsUploadFlavor == "" {
		return multistep.ActionContinue
	}

	path := driver.ToolsIsoPath(c.ToolsUploadFlavor)
	if _, err := os.Stat(path); err != nil {
		state.Put("error", fmt.Errorf(
			"Couldn't find VMware tools for '%s'! VMware often downloads these\n"+
				"tools on-demand. However, to do this, you need to create a fake VM\n"+
				"of the proper type then click the 'install tools' option in the\n"+
				"VMware GUI.", c.ToolsUploadFlavor))
		return multistep.ActionHalt
	}

	state.Put("tools_upload_source", path)
	return multistep.ActionContinue
}

func (c *StepPrepareTools) Cleanup(multistep.StateBag) {}
