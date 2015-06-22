package chroot

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"os"
	"os/exec"
	"syscall"
)

// StepMountExtra mounts the attached device.
//
// Produces:
//   mount_extra_cleanup CleanupFunc - To perform early cleanup
type StepMountExtra struct {
	mounts []string
}

func (s *StepMountExtra) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	mountPath := state.Get("mount_path").(string)
	ui := state.Get("ui").(packer.Ui)
	wrappedCommand := state.Get("wrappedCommand").(CommandWrapper)

	s.mounts = make([]string, 0, len(config.ChrootMounts))

	ui.Say("Mounting additional paths within the chroot...")
	for _, mountInfo := range config.ChrootMounts {
		innerPath := mountPath + mountInfo[2]

		if err := os.MkdirAll(innerPath, 0755); err != nil {
			err := fmt.Errorf("Error creating mount directory: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		flags := "-t " + mountInfo[0]
		if mountInfo[0] == "bind" {
			flags = "--bind"
		}

		ui.Message(fmt.Sprintf("Mounting: %s", mountInfo[2]))
		stderr := new(bytes.Buffer)
		mountCommand, err := wrappedCommand(fmt.Sprintf(
			"mount %s %s %s",
			flags,
			mountInfo[1],
			innerPath))
		if err != nil {
			err := fmt.Errorf("Error creating mount command: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		cmd := ShellCommand(mountCommand)
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			err := fmt.Errorf(
				"Error mounting: %s\nStderr: %s", err, stderr.String())
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		s.mounts = append(s.mounts, innerPath)
	}

	state.Put("mount_extra_cleanup", s)
	return multistep.ActionContinue
}

func (s *StepMountExtra) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)

	if err := s.CleanupFunc(state); err != nil {
		ui.Error(err.Error())
		return
	}
}

func (s *StepMountExtra) CleanupFunc(state multistep.StateBag) error {
	if s.mounts == nil {
		return nil
	}

	wrappedCommand := state.Get("wrappedCommand").(CommandWrapper)
	for len(s.mounts) > 0 {
		var path string
		lastIndex := len(s.mounts) - 1
		path, s.mounts = s.mounts[lastIndex], s.mounts[:lastIndex]

		grepCommand, err := wrappedCommand(fmt.Sprintf("grep %s /proc/mounts", path))
		if err != nil {
			return fmt.Errorf("Error creating grep command: %s", err)
		}

		// Before attempting to unmount,
		// check to see if path is already unmounted
		stderr := new(bytes.Buffer)
		cmd := ShellCommand(grepCommand)
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
					exitStatus := status.ExitStatus()
					if exitStatus == 1 {
						// path has already been unmounted
						// just skip this path
						continue
					}
				}
			}
		}

		unmountCommand, err := wrappedCommand(fmt.Sprintf("umount %s", path))
		if err != nil {
			return fmt.Errorf("Error creating unmount command: %s", err)
		}

		stderr = new(bytes.Buffer)
		cmd = ShellCommand(unmountCommand)
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf(
				"Error unmounting device: %s\nStderr: %s", err, stderr.String())
		}
	}

	s.mounts = nil
	return nil
}
