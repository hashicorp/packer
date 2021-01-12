package chroot

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
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
	MountPartition string

	mountPath     string
	GeneratedData *packerbuilderdata.GeneratedData
}

func (s *StepMountDevice) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)
	device := state.Get("device").(string)
	if config.NVMEDevicePath != "" {
		// customizable device path for mounting NVME block devices on c5 and m5 HVM
		device = config.NVMEDevicePath
	}
	wrappedCommand := state.Get("wrappedCommand").(common.CommandWrapper)

	var virtualizationType string
	if config.FromScratch || config.AMIVirtType != "" {
		virtualizationType = config.AMIVirtType
	} else {
		image := state.Get("source_image").(*ec2.Image)
		virtualizationType = *image.VirtualizationType
		log.Printf("Source image virtualization type is: %s", virtualizationType)
	}

	ictx := config.ctx

	ictx.Data = &mountPathData{Device: filepath.Base(device)}
	mountPath, err := interpolate.Render(config.MountPath, &ictx)

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

	deviceMount := device

	if virtualizationType == "hvm" && s.MountPartition != "0" {
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
		err := fmt.Errorf("Error creating mount command: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	log.Printf("[DEBUG] (step mount) mount command is %s", mountCommand)
	cmd := common.ShellCommand(mountCommand)
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
	s.GeneratedData.Put("MountPath", s.mountPath)
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
		return fmt.Errorf("Error creating unmount command: %s", err)
	}

	cmd := common.ShellCommand(unmountCommand)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Error unmounting root device: %s", err)
	}

	s.mountPath = ""
	return nil
}
