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
//   mount_path string - The location where the volume was mounted.
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

	return multistep.ActionContinue
}

func (s *StepMountExtra) Cleanup(state map[string]interface{}) {
	if s.mounts == nil {
		return
	}

	config := state["config"].(*Config)
	ui := state["ui"].(packer.Ui)

	for i := len(s.mounts) - 1; i >= 0; i-- {
		path := s.mounts[i]
		unmountCommand := fmt.Sprintf("%s %s", config.UnmountCommand, path)

		stderr := new(bytes.Buffer)
		cmd := exec.Command("/bin/sh", "-c", unmountCommand)
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			ui.Error(fmt.Sprintf(
				"Error unmounting device: %s\nStderr: %s", err, stderr.String()))
			return
		}
	}
}
