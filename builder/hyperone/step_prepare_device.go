package hyperone

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

const (
	vmBusPath = "/sys/bus/vmbus/devices"
)

type stepPrepareDevice struct{}

func (s *stepPrepareDevice) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	config := state.Get("config").(*Config)

	if config.ChrootDevice != "" {
		state.Put("device", config.ChrootDevice)
		return multistep.ActionContinue
	}

	controllerNumber := state.Get("chroot_controller_number").(string)
	controllerLocation := state.Get("chroot_controller_location").(int)

	log.Println("Searching for available device...")

	cmd := fmt.Sprintf("find %s/%s/ -path *:%d/block -exec ls {} \\;",
		vmBusPath, controllerNumber, controllerLocation)

	block, err := captureOutput(cmd, state)
	if err != nil {
		err := fmt.Errorf("error finding available device: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if block == "" {
		err := fmt.Errorf("device not found")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	device := fmt.Sprintf("/dev/%s", block)

	ui.Say(fmt.Sprintf("Found device: %s", device))
	state.Put("device", device)
	return multistep.ActionContinue
}

func (s *stepPrepareDevice) Cleanup(state multistep.StateBag) {}
