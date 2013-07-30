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
	"text/template"
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

func (s *StepMountDevice) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*Config)
	ui := state["ui"].(packer.Ui)
	device := state["device"].(string)

	mountPathRaw := new(bytes.Buffer)
	t := template.Must(template.New("mountPath").Parse(config.MountPath))
	t.Execute(mountPathRaw, &mountPathData{
		Device: filepath.Base(device),
	})

	mountPath := mountPathRaw.String()
	log.Printf("Mount path: %s", mountPath)

	if err := os.MkdirAll(mountPath, 0755); err != nil {
		err := fmt.Errorf("Error creating mount directory: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("Mounting the root device...")
	stderr := new(bytes.Buffer)
	mountCommand := fmt.Sprintf("%s %s %s", config.MountCommand, device, mountPath)
	cmd := exec.Command("/bin/sh", "-c", mountCommand)
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		err := fmt.Errorf(
			"Error mounting root volume: %s\nStderr: %s", err, stderr.String())
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the mount path so we remember to unmount it later
	s.mountPath = mountPath
	state["mount_path"] = s.mountPath
	state["mount_device_cleanup"] = s.CleanupFunc

	return multistep.ActionContinue
}

func (s *StepMountDevice) Cleanup(state map[string]interface{}) {
	ui := state["ui"].(packer.Ui)
	if err := s.CleanupFunc(state); err != nil {
		ui.Error(err.Error())
	}
}

func (s *StepMountDevice) CleanupFunc(state map[string]interface{}) error {
	if s.mountPath == "" {
		return nil
	}

	config := state["config"].(*Config)
	ui := state["ui"].(packer.Ui)
	ui.Say("Unmounting the root device...")

	unmountCommand := fmt.Sprintf("%s %s", config.UnmountCommand, s.mountPath)
	cmd := exec.Command("/bin/sh", "-c", unmountCommand)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error unmounting root device: %s", err)
	}

	s.mountPath = ""
	return nil
}
