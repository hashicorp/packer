package chroot

// mostly borrowed from ./builder/amazon/chroot/step_mount_device.go

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

var _ multistep.Step = &StepMountDevice{}

type StepMountDevice struct {
	MountOptions   []string
	MountPartition string
	MountPath      string

	mountPath string
}

func (s *StepMountDevice) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	device := state.Get("device").(string)
	config := state.Get("config").(*Config)
	wrappedCommand := state.Get("wrappedCommand").(common.CommandWrapper)

	ictx := config.ctx

	ictx.Data = &struct{ Device string }{Device: filepath.Base(device)}
	mountPath, err := interpolate.Render(s.MountPath, &ictx)

	if err != nil {
		err := fmt.Errorf("error preparing mount directory: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	mountPath, err = filepath.Abs(mountPath)
	if err != nil {
		err := fmt.Errorf("error preparing mount directory: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Mount path: %s", mountPath)

	if err := os.MkdirAll(mountPath, 0755); err != nil {
		err := fmt.Errorf("error creating mount directory: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	var deviceMount string
	switch runtime.GOOS {
	case "freebsd":
		deviceMount = fmt.Sprintf("%sp%s", device, s.MountPartition)
	default:
		deviceMount = fmt.Sprintf("%s%s", device, s.MountPartition)
	}

	state.Put("deviceMount", deviceMount)

	ui.Say("Mounting the root device...")
	stderr := new(bytes.Buffer)

	// build mount options from mount_options config, useful for nouuid options
	// or other specific device type settings for mount
	opts := ""
	if len(s.MountOptions) > 0 {
		opts = "-o " + strings.Join(s.MountOptions, " -o ")
	}
	mountCommand, err := wrappedCommand(
		fmt.Sprintf("mount %s %s %s", opts, deviceMount, mountPath))
	if err != nil {
		err := fmt.Errorf("error creating mount command: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	log.Printf("[DEBUG] (step mount) mount command is %s", mountCommand)
	cmd := common.ShellCommand(mountCommand)
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		err := fmt.Errorf(
			"error mounting root volume: %s\nStderr: %s", err, stderr.String())
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
	ui := state.Get("ui").(packersdk.Ui)
	if err := s.CleanupFunc(state); err != nil {
		ui.Error(err.Error())
	}
}

func (s *StepMountDevice) CleanupFunc(state multistep.StateBag) error {
	if s.mountPath == "" {
		return nil
	}

	ui := state.Get("ui").(packersdk.Ui)
	wrappedCommand := state.Get("wrappedCommand").(common.CommandWrapper)

	ui.Say("Unmounting the root device...")
	unmountCommand, err := wrappedCommand(fmt.Sprintf("umount %s", s.mountPath))
	if err != nil {
		return fmt.Errorf("error creating unmount command: %s", err)
	}

	cmd := common.ShellCommand(unmountCommand)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error unmounting root device: %s", err)
	}

	s.mountPath = ""
	return nil
}
