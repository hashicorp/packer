package vmware

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"os/exec"
	"path/filepath"
)

type stepCreateDisk struct{}

func (stepCreateDisk) Run(state map[string]interface{}) multistep.StepAction {
	// TODO(mitchellh): Configurable disk size
	// TODO(mitchellh): Capture error output in case things go wrong to report it

	config := state["config"].(*config)
	ui := state["ui"].(packer.Ui)

	vdisk_manager := "/Applications/VMware Fusion.app/Contents/Library/vmware-vdiskmanager"
	output := filepath.Join(config.OutputDir, "disk.vmdk")

	ui.Say("Creating virtual machine disk")
	cmd := exec.Command(vdisk_manager, "-c", "-s", "40000M", "-a", "lsilogic", "-t", "1", output)
	if err := cmd.Run(); err != nil {
		ui.Error(fmt.Sprintf("Error creating VMware disk: %s", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (stepCreateDisk) Cleanup(map[string]interface{}) {}
