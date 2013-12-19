package virtualbox

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step removes any devices (floppy disks, ISOs, etc.) from the
// machine that we may have added.
//
// Uses:
//
// Produces:
type stepRemoveDevices struct{}

func (s *stepRemoveDevices) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	// Remove the attached floppy disk, if it exists
	if _, ok := state.GetOk("floppy_path"); ok {
		ui.Message("Removing floppy drive...")
		command := []string{
			"storageattach", vmName,
			"--storagectl", "Floppy Controller",
			"--port", "0",
			"--device", "0",
			"--medium", "none",
		}
		if err := driver.VBoxManage(command...); err != nil {
			err := fmt.Errorf("Error removing floppy: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	command := []string{
		"storageattach", vmName,
		"--storagectl", "IDE Controller",
		"--port", "0",
		"--device", "1",
		"--medium", "none",
	}

	if err := driver.VBoxManage(command...); err != nil {
		err := fmt.Errorf("Error detaching ISO: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepRemoveDevices) Cleanup(state multistep.StateBag) {
}
