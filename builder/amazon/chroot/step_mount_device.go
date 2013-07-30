package chroot

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"os"
	"os/exec"
)

// StepMountDevice mounts the attached device.
//
// Produces:
//   mount_path string - The location where the volume was mounted.
type StepMountDevice struct {
	mountPath string
}

func (s *StepMountDevice) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*Config)
	ui := state["ui"].(packer.Ui)
	device := state["device"].(string)

	mountPath := config.MountPath
	log.Printf("Mount path: %s", mountPath)

	if err := os.MkdirAll(mountPath, 0755); err != nil {
		err := fmt.Errorf("Error creating mount directory: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Mounting the root device...")
	stderr := new(bytes.Buffer)
	cmd := exec.Command("mount", device, mountPath)
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		err := fmt.Errorf(
			"Error mounting root volume: %s\nStderr: %s", err, stderr.String())
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepMountDevice) Cleanup(state map[string]interface{}) {
	if s.mountPath == "" {
		return
	}

	ui := state["ui"].(packer.Ui)
	ui.Say("Unmounting the root device...")

	path, err := exec.LookPath("umount")
	if err != nil {
		ui.Error(fmt.Sprintf("Error umounting root device: %s", err))
		return
	}

	cmd := exec.Command(path, s.mountPath)
	if err := cmd.Run(); err != nil {
		ui.Error(fmt.Sprintf(
			"Error unmounting root device: %s", err))
		return
	}
}
