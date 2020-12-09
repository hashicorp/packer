package chroot

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepPrepareDevice finds an available device and sets it.
type StepPrepareDevice struct {
}

func (s *StepPrepareDevice) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

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
