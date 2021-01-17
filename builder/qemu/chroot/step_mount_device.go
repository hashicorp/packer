package chroot

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// StepMountDevice mounts the mapped device.
//
// Produces:
//   mount_path string - The location where the device was mounted.
//   mount_device_cleanup CleanupFunc - To perform early cleanup
type mountPathData struct {
	Device string
}

type StepMountDevice struct {
	MountOptions []string

	mountPath     string
	GeneratedData *packerbuilderdata.GeneratedData
}

func (s *StepMountDevice) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)
	device := state.Get("device").(string)

	ctx := config.ctx
	ctx.Data = &mountPathData{Device: filepath.Base(device)}

	mountPath, err := interpolate.Render(config.MountPath, &ctx)
	if err != nil {
		return Halt(state, fmt.Errorf("failed to prepare mount directory: %s", err))
	}

	mountPath, err = filepath.Abs(mountPath)
	if err != nil {
		return Halt(state, fmt.Errorf("failed to prepare mount directory: %s", err))
	}

	log.Printf("Mount path: %s", mountPath)

	if err := os.MkdirAll(mountPath, 0755); err != nil {
		return Halt(state, fmt.Errorf("Failed to create mount directory \"%s\": %s", mountPath, err))
	}

	ui.Say("Mounting device...")

	opts := ""
	if len(s.MountOptions) > 0 {
		opts = "-o " + strings.Join(s.MountOptions, " -o ")
	}

	if _, err := RunCommand(state, fmt.Sprintf("mount %s %s %s", opts, device, mountPath)); err != nil {
		return Halt(state, fmt.Errorf("Cannot mount device \"%s\": %s", device, err))
	}

	// Set the mount path so we remember to unmount it later
	s.mountPath = mountPath
	state.Put("mount_path", s.mountPath)
	s.GeneratedData.Put("MountPath", s.mountPath)
	state.Put("mount_device_cleanup", s)

	return multistep.ActionContinue
}

func (s *StepMountDevice) CleanupFunc(state multistep.StateBag) error {
	ui := state.Get("ui").(packersdk.Ui)

	if s.mountPath == "" {
		return nil
	}

	ui.Say(fmt.Sprintf("Unmounting device \"%s\"...", s.mountPath))

	if _, err := RunCommand(state, fmt.Sprintf("umount %s", s.mountPath)); err != nil {
		ui.Error(fmt.Sprintf("Failed to unmount device: %s", err))
	} else {
		s.mountPath = ""
	}

	return nil
}

func (s *StepMountDevice) Cleanup(state multistep.StateBag) {
	_ = s.CleanupFunc(state)
}
