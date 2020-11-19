package googlecompute

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// StepTeardownInstance represents a Packer build step that tears down GCE
// instances.
type StepTeardownInstance struct {
	Debug bool
}

// Run executes the Packer build step that tears down a GCE instance.
func (s *StepTeardownInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	name := config.InstanceName
	if name == "" {
		return multistep.ActionHalt
	}

	ui.Say("Deleting instance...")
	instanceLog, _ := driver.GetSerialPortOutput(config.Zone, name)
	state.Put("instance_log", instanceLog)
	errCh, err := driver.DeleteInstance(config.Zone, name)
	if err == nil {
		select {
		case err = <-errCh:
		case <-time.After(config.StateTimeout):
			err = errors.New("time out while waiting for instance to delete")
		}
	}

	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error deleting instance. Please delete it manually.\n\n"+
				"Name: %s\n"+
				"Error: %s", name, err))
		return multistep.ActionHalt
	}
	ui.Message("Instance has been deleted!")
	state.Put("instance_name", "")

	return multistep.ActionContinue
}

// Deleting the instance does not remove the boot disk. This cleanup removes
// the disk.
func (s *StepTeardownInstance) Cleanup(state multistep.StateBag) {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Deleting disk...")
	errCh, err := driver.DeleteDisk(config.Zone, config.DiskName)
	if err == nil {
		select {
		case err = <-errCh:
		case <-time.After(config.StateTimeout):
			err = errors.New("time out while waiting for disk to delete")
		}
	}

	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error deleting disk. Please delete it manually.\n\n"+
				"DiskName: %s\n"+
				"Zone: %s\n"+
				"Error: %s", config.DiskName, config.Zone, err))
	}

	ui.Message("Disk has been deleted!")

	return
}
