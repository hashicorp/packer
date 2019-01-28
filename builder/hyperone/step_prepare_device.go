package hyperone

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepPrepareDevice struct{}

func (s *stepPrepareDevice) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	chrootDiskID := state.Get("chroot_disk_id").(string)

	var err error
	log.Println("Searching for available device...")
	device, err := availableDevice(chrootDiskID)
	if err != nil {
		err := fmt.Errorf("error finding available device: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if _, err := os.Stat(device); err == nil {
		err := fmt.Errorf("device is in use: %s", device)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Found device: %s", device))
	state.Put("device", device)
	return multistep.ActionContinue
}

func (s *stepPrepareDevice) Cleanup(state multistep.StateBag) {}

func availableDevice(scsiID string) (string, error) {
	// TODO proper SCSI search
	return "/dev/sdb", nil
}
