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

		ui.Message(fmt.Sprintf("Mounting: %s", mountInfo[2]))
		stderr := new(bytes.Buffer)
		mountCommand := fmt.Sprintf(
			"%s -t %s %s %s",
			config.MountCommand,
			mountInfo[0],
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

	for _, path := range s.mounts {
		unmountCommand := fmt.Sprintf("%s %s", config.UnmountCommand, path)
		cmd := exec.Command("bin/sh", "-c", unmountCommand)
		if err := cmd.Run(); err != nil {
			ui.Error(fmt.Sprintf(
				"Error unmounting root device: %s", err))
			return
		}
	}
}
