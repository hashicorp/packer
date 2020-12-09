package iso

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	parallelscommon "github.com/hashicorp/packer/builder/parallels/common"
)

// This step sets the device boot order for the virtual machine.
//
// Uses:
//   driver Driver
//   ui packersdk.Ui
//   vmName string
//
// Produces:
type stepSetBootOrder struct{}

func (s *stepSetBootOrder) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(parallelscommon.Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmName := state.Get("vmName").(string)

	// Set new boot order
	ui.Say("Setting the boot order...")
	command := []string{
		"set", vmName,
		"--device-bootorder", fmt.Sprintf("hdd0 cdrom0 net0"),
	}

	if err := driver.Prlctl(command...); err != nil {
		err := fmt.Errorf("Error setting the boot order: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepSetBootOrder) Cleanup(state multistep.StateBag) {}
