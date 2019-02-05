package hyperone

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

const (
	diskByPathPrefix = "/dev/disk/by-path/acpi-VMBUS:01-scsi-0:0:0:"
)

type stepPrepareDevice struct{}

func (s *stepPrepareDevice) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	chrootDiskLocation := state.Get("chroot_disk_location").(int)

	log.Println("Searching for available device...")

	diskByPath := fmt.Sprintf("%s%d", diskByPathPrefix, chrootDiskLocation)
	cmd := fmt.Sprintf("readlink -f %s", diskByPath)

	device, err := captureOutput(cmd, state)
	if err != nil {
		err := fmt.Errorf("error finding available device: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Found device: %s -> %s", diskByPath, device))
	state.Put("device", device)
	return multistep.ActionContinue
}

func (s *stepPrepareDevice) Cleanup(state multistep.StateBag) {}
