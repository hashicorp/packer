package surrogate

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/packer/builder/google/gcp"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepTeardownInstance represents a Packer build step that tears down GCE
// instances.
type StepTeardownInstance struct {
	Debug bool
}

// Run executes the Packer build step that tears down a GCE instance.
func (s *StepTeardownInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(gcp.Driver)
	ui := state.Get("ui").(packer.Ui)

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
	driver := state.Get("driver").(gcp.Driver)
	ui := state.Get("ui").(packer.Ui)

	ui.Message(fmt.Sprintf("Boot %+v", config.BootDisk))
	ui.Message(fmt.Sprintf("Target %+v", config.TargetDisk))

	deleteDisk(ui, driver, "Boot", config.BootDisk.Name, config.Zone, config.StateTimeout)
	deleteDisk(ui, driver, "Target", config.TargetDisk.Name, config.Zone, config.StateTimeout)
}

// deleteDisk deletes a disk and displays UI messages describing progress.
func deleteDisk(ui packer.Ui, driver gcp.Driver, description string, diskName string, zone string, timeout time.Duration) {
	ui.Say(fmt.Sprintf("Deleting %s disk...", strings.ToLower(description)))
	errCh, err := driver.DeleteDisk(zone, diskName)
	if err == nil {
		select {
		case err = <-errCh:
		case <-time.After(timeout):
			err = errors.New("time out while waiting for disk to delete")
		}
	}
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error deleting disk. Please delete it manually.\n\n"+
				"DiskName: %s\n"+
				"Zone: %s\n"+
				"Error: %s", diskName, zone, err))
	}

	ui.Message(fmt.Sprintf("%s disk has been deleted!", description))
}
