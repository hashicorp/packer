package iso

import (
	"fmt"
	"github.com/mitchellh/multistep"
	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
	"os"
)

type stepPrepareTools struct{}

func (*stepPrepareTools) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	driver := state.Get("driver").(vmwcommon.Driver)

	if config.ToolsUploadFlavor == "" {
		return multistep.ActionContinue
	}

	path := driver.ToolsIsoPath(config.ToolsUploadFlavor)
	if _, err := os.Stat(path); err != nil {
		state.Put("error", fmt.Errorf(
			"Couldn't find VMware tools for '%s'! VMware often downloads these\n"+
				"tools on-demand. However, to do this, you need to create a fake VM\n"+
				"of the proper type then click the 'install tools' option in the\n"+
				"VMware GUI.", config.ToolsUploadFlavor))
		return multistep.ActionHalt
	}

	state.Put("tools_upload_source", path)
	return multistep.ActionContinue
}

func (*stepPrepareTools) Cleanup(multistep.StateBag) {}
