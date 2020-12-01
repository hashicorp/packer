package iso

import (
	"context"
	"fmt"
	"log"

	parallelscommon "github.com/hashicorp/packer/builder/parallels/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// This step attaches the ISO to the virtual machine.
//
// Uses:
//   driver Driver
//   iso_path string
//   ui packersdk.Ui
//   vmName string
//
// Produces:
//	 attachedIso bool
type stepAttachISO struct{}

func (s *stepAttachISO) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(parallelscommon.Driver)
	isoPath := state.Get("iso_path").(string)
	ui := state.Get("ui").(packersdk.Ui)
	vmName := state.Get("vmName").(string)

	// Attach the disk to the cdrom0 device. We couldn't use a separated device because it is failed to boot in PD9 [GH-1667]
	ui.Say("Attaching ISO to the default CD/DVD ROM device...")
	command := []string{
		"set", vmName,
		"--device-set", "cdrom0",
		"--image", isoPath,
		"--enable", "--connect",
	}
	if err := driver.Prlctl(command...); err != nil {
		err := fmt.Errorf("Error attaching ISO: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set some state so we know to remove
	state.Put("attachedIso", true)

	return multistep.ActionContinue
}

func (s *stepAttachISO) Cleanup(state multistep.StateBag) {
	if _, ok := state.GetOk("attachedIso"); !ok {
		return
	}

	driver := state.Get("driver").(parallelscommon.Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmName := state.Get("vmName").(string)

	// Detach ISO by setting an empty string image.
	log.Println("Detaching ISO from the default CD/DVD ROM device...")
	command := []string{
		"set", vmName,
		"--device-set", "cdrom0",
		"--image", "", "--disconnect", "--enable",
	}

	if err := driver.Prlctl(command...); err != nil {
		ui.Error(fmt.Sprintf("Error detaching ISO: %s", err))
	}
}
