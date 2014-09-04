package chroot

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"os"
)

// StepPrepareDevice finds an available device and sets it.
type StepPrepareDevice struct {
	mounts []string
}

func (s *StepPrepareDevice) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	device := config.DevicePath
	if device == "" {
		var err error
		log.Println("Device path not specified, searching for available device...")
		device, err = AvailableDevice()
		if err != nil {
			err := fmt.Errorf("Error finding available device: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	if _, err := os.Stat(device); err == nil {
		err := fmt.Errorf("Device is in use: %s", device)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Device: %s", device)
	state.Put("device", device)
	return multistep.ActionContinue
}

func (s *StepPrepareDevice) Cleanup(state multistep.StateBag) {}
