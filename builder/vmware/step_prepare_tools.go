package vmware

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"os"
)

type stepPrepareTools struct{}

func (*stepPrepareTools) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	driver := state["driver"].(Driver)

	if config.ToolsUploadFlavor == "" {
		return multistep.ActionContinue
	}

	path := driver.ToolsIsoPath(config.ToolsUploadFlavor)
	if _, err := os.Stat(path); err != nil {
		state["error"] = fmt.Errorf(
			"Couldn't find VMware tools for '%s'! VMware often downloads these\n"+
				"tools on-demand. However, to do this, you need to create a fake VM\n"+
				"of the proper type then click the 'install tools' option in the\n"+
				"VMware GUI.", config.ToolsUploadFlavor)
		return multistep.ActionHalt
	}

	state["tools_upload_source"] = path
	return multistep.ActionContinue
}

func (*stepPrepareTools) Cleanup(map[string]interface{}) {}
