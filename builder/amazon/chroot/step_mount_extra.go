package chroot

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"os"
	"os/exec"
)

// StepMountExtra mounts the attached device.
//
// Produces:
//   mount_extra_cleanup CleanupFunc - To perform early cleanup
type StepMountExtra struct {
	mounts []string
}

func (s *StepMountExtra) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*Config)
	mountPath := state["mount_path"].(string)
	ui := state["ui"].(packer.Ui)

	s.mounts = make([]string, 0, len(config.ChrootMounts))

	ui.Say("Mounting additional paths within the chroot...")
	for _, mountInfo := range config.ChrootMounts {
		innerPath := mountPath + mountInfo[2]

		if err := os.MkdirAll(innerPath, 0755); err != nil {
			err := fmt.Errorf("Error creating mount directory: %s", err)
			state["error"] = err
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		flags := "-t " + mountInfo[0]
		if mountInfo[0] == "bind" {
			flags = "--bind"
		}

		ui.Message(fmt.Sprintf("Mounting: %s", mountInfo[2]))
		stderr := new(bytes.Buffer)
		mountCommand := fmt.Sprintf(
			"%s %s %s %s",
			config.MountCommand,
			flags,
			mountInfo[1],
			innerPath)
		cmd := exec.Command("/bin/sh", "-c", mountCommand)
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			err := fmt.Errorf(
				"Error mounting: %s\nStderr: %s", err, stderr.String())
			state["error"] = err
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		s.mounts = append(s.mounts, innerPath)
	}

	state["mount_extra_cleanup"] = s.CleanupFunc
	return multistep.ActionContinue
}

func (s *StepMountExtra) Cleanup(state map[string]interface{}) {
	ui := state["ui"].(packer.Ui)

	if err := s.CleanupFunc(state); err != nil {
		ui.Error(err.Error())
		return
	}
}

func (s *StepMountExtra) CleanupFunc(state map[string]interface{}) error {
	if s.mounts == nil {
		return nil
	}

	config := state["config"].(*Config)
	for len(s.mounts) > 0 {
		var path string
		lastIndex := len(s.mounts) - 1
		path, s.mounts = s.mounts[lastIndex], s.mounts[:lastIndex]
		unmountCommand := fmt.Sprintf("%s %s", config.UnmountCommand, path)

		stderr := new(bytes.Buffer)
		cmd := exec.Command("/bin/sh", "-c", unmountCommand)
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf(
				"Error unmounting device: %s\nStderr: %s", err, stderr.String())
		}
	}

	s.mounts = nil
	return nil
}
