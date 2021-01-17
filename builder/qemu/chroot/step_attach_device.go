package chroot

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
)

// StepAttachVolume mapping source image to device
type StepAttachVolume struct {
	GeneratedData *packerbuilderdata.GeneratedData

	MountPartition int

	device       string
	rawImage     string
	numPartition string
	deviceMapped bool
}

func (s *StepAttachVolume) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	s.rawImage = state.Get("rawImage").(string)

	var err error
	// check if image contain partition
	if s.numPartition, err = RunCommand(state, fmt.Sprintf("set -o pipefail >/dev/null 2>&1; kpartx -l %s | awk \"/loop[0-9]+p/\"|wc -l", s.rawImage)); err != nil {
		return Halt(state, fmt.Errorf("Failed to check if source image has partitions: %s", err))
	}

	var numPartition int
	if numPartition, err = strconv.Atoi(s.numPartition); err != nil {
		return Halt(state, fmt.Errorf("Cannot get number of partitions of source image %s", err))
	}

	if s.MountPartition > numPartition {
		return Halt(state, fmt.Errorf("MountPartition %d does not exist", s.MountPartition))
	}

	if _, err := RunCommand(state, fmt.Sprintf("udevadm settle")); err != nil {
		return Halt(state, fmt.Errorf("Failed to wait for udev ready: %s", err))
	}

	if _, err := RunCommand(state, fmt.Sprintf("losetup -f")); err != nil {
		return Halt(state, fmt.Errorf("No loop device is available: %s", err))
	}

	s.deviceMapped = true

	// add partition mapping
	cmd := fmt.Sprintf("set -o pipefail >/dev/null 2>&1; kpartx -av %s | awk \"/loop[0-9]+p%d/ {print \\$3}\"", s.rawImage, s.MountPartition)
	if s.device, err = RunCommand(state, cmd); err != nil {
		return Halt(state, fmt.Errorf("error connecting to the source image: %s", err))
	}

	if s.device == "" {
		return Halt(state, fmt.Errorf("Failed to create image device"))
	}

	// Wait for the device to be connected.
	time.Sleep(5 * time.Second)

	s.device = fmt.Sprintf("/dev/mapper/%s", s.device)

	if _, err := os.Stat(s.device); os.IsNotExist(err) {
		return Halt(state, fmt.Errorf("Failed to map device: %s", err))
	}

	ui := state.Get("ui").(packersdk.Ui)
	ui.Say(fmt.Sprintf("Device is ready \"%s\"", s.device))

	state.Put("attach_cleanup", s)
	state.Put("device", s.device)
	s.GeneratedData.Put("device", s.device)

	return multistep.ActionContinue
}

func (s *StepAttachVolume) Cleanup(state multistep.StateBag) {
	_ = s.CleanupFunc(state)
}

func (s *StepAttachVolume) CleanupFunc(state multistep.StateBag) error {
	ui := state.Get("ui").(packersdk.Ui)

	if !s.deviceMapped {
		return nil
	}

	ui.Say("Unmapping the source image...")

	if _, err := RunCommand(state, fmt.Sprintf("kpartx -d %s", s.rawImage)); err != nil {
		ui.Error(fmt.Sprintf("Failed to ummapping device \"%s\"", s.device))
	} else {
		s.deviceMapped = false
	}

	return nil
}
