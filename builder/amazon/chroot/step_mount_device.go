package chroot

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
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
	MountOptions   []string
	MountPartition int

	mountPath string
}

func (s *StepMountDevice) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	image := state.Get("source_image").(*ec2.Image)
	device := state.Get("device").(string)
	wrappedCommand := state.Get("wrappedCommand").(CommandWrapper)

	ctx := config.ctx
	ctx.Data = &mountPathData{Device: filepath.Base(device)}
	mountPath, err := interpolate.Render(config.MountPath, &ctx)

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

	log.Printf("Source image virtualization type is: %s", *image.VirtualizationType)
	deviceMount := device
	if *image.VirtualizationType == "hvm" {
		deviceMount = fmt.Sprintf("%s%d", device, s.MountPartition)
	}
	state.Put("deviceMount", deviceMount)

	ui.Say("Mounting the root device...")
	stderr := new(bytes.Buffer)

	// build mount options from mount_options config, usefull for nouuid options
	// or other specific device type settings for mount
	opts := ""
	if len(s.MountOptions) > 0 {
		opts = "-o " + strings.Join(s.MountOptions, " -o ")
	}
	mountCommand, err := wrappedCommand(
		fmt.Sprintf("mount %s %s %s", opts, deviceMount, mountPath))
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
	wrappedCommand := state.Get("wrappedCommand").(CommandWrapper)

	ui.Say("Unmounting the root device...")
	unmountCommand, err := wrappedCommand(fmt.Sprintf("umount %s", s.mountPath))
	if err != nil {
		return fmt.Errorf("Error creating unmount command: %s", err)
	}

	cmd := ShellCommand(unmountCommand)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error unmounting root device: %s", err)
	}

	s.mountPath = ""
	return nil
}
