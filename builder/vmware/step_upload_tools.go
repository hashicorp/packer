package vmware

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"os"
	"fmt"
)

type stepUploadTools struct{}

func (*stepUploadTools) Run(state map[string]interface{}) multistep.StepAction {
	comm := state["communicator"].(packer.Communicator)
	ui := state["ui"].(packer.Ui)
	driver := state["driver"].(Driver)

	ui.Say("Uploading the VMware Tools.")

	f, err := os.Open(driver.ToolsIsoPath())
	if err != nil {
		state["error"] = fmt.Errorf("Error opening VMware Tools ISO: %s", err)
		return multistep.ActionHalt
	}
	defer f.Close()

	if err := comm.Upload("/tmp/linux.iso", f); err != nil {
		state["error"] = fmt.Errorf("Error uploading VMware Tools: %s", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (*stepUploadTools) Cleanup(map[string]interface{}) {}
