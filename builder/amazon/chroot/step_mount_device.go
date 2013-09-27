package chroot

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type mountPathData struct {
	Device string
}

// StepMountDevice mounts the attached device.
//
// Produces:
//   mount_path string - The location where the volume was mounted.
//   mount_device_cleanup CleanupFunc - To perform early cleanup
type StepMountDevice struct {
	mountPath string
}

func (s *StepMountDevice) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	device := state.Get("device").(string)

	mountPath, err := config.tpl.Process(config.MountPath, &mountPathData{
		Device: filepath.Base(device),
	})

	if err != nil {
		err := fmt.Errorf("Error preparing mount directory: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	mountPath, err = filepath.Abs(mountPath)
	if err != nil {
		err := fmt.Errorf("Error preparing mount directory: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Mount path: %s", mountPath)

	if err := os.MkdirAll(mountPath, 0755); err != nil {
		err := fmt.Errorf("Error creating mount directory: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	chrootCmd := func(command string) *exec.Cmd {
		return ChrootCommand(mountPath, command)
	}
	state.Put("chrootCmd", Command(chrootCmd))

	ui.Say("Mounting the root device...")
	stderr := new(bytes.Buffer)
	mountCommand := fmt.Sprintf("mount %s %s", device, mountPath)
	wrappedCommand := state.Get("wrappedCommand").(Command)
	cmd := wrappedCommand(mountCommand)
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		err := fmt.Errorf(
			"Error mounting root volume: %s\nStderr: %s", err, stderr.String())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the mount path so we remember to unmount it later
	s.mountPath = mountPath
	state.Put("mount_path", s.mountPath)
	state.Put("mount_device_cleanup", s)

	return multistep.ActionContinue
}

func (s *StepMountDevice) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	if err := s.CleanupFunc(state); err != nil {
		ui.Error(err.Error())
	}
}

func (s *StepMountDevice) CleanupFunc(state multistep.StateBag) error {
	if s.mountPath == "" {
		return nil
	}

	ui := state.Get("ui").(packer.Ui)
	ui.Say("Unmounting the root device...")

	unmountCommand := fmt.Sprintf("umount %s", s.mountPath)
	wrappedCommand := state.Get("wrappedCommand").(Command)
	cmd := wrappedCommand(unmountCommand)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error unmounting root device: %s", err)
	}

	s.mountPath = ""
	return nil
}
