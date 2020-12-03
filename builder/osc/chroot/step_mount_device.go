package chroot

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
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

	mountPath string
}

func (s *StepMountDevice) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)
	device := state.Get("device").(string)
	if config.NVMEDevicePath != "" {
		// customizable device path for mounting NVME block devices on c5 and m5 HVM
		device = config.NVMEDevicePath
	}
	wrappedCommand := state.Get("wrappedCommand").(CommandWrapper)

	var virtualizationType string
	if config.FromScratch {
		virtualizationType = config.OMIVirtType
	} else {
		//image := state.Get("source_image").(oapi.Image)

		//Is always hvm
		virtualizationType = "hvm"
		log.Printf("Source image virtualization type is: %s", virtualizationType)
	}

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

	log.Printf("[DEBUG] Device: %s", device)
	log.Printf("[DEBUG] Mount path: %s", mountPath)

	if err := os.MkdirAll(mountPath, 0755); err != nil {
		err := fmt.Errorf("Error creating mount directory: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	//Check the symbolic link for the device to get the real device name
	cmd := ShellCommand(fmt.Sprintf("lsblk -no pkname $(readlink -f %s)", device))

	realDeviceName, err := cmd.Output()
	if err != nil {
		err := fmt.Errorf(
			"Error retrieving the symlink of the device %s.\n", device)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("[DEBUG] RealDeviceName: %s", realDeviceName)

	realDeviceNameSplitted := strings.Split(string(realDeviceName), "\n")
	log.Printf("[DEBUG] RealDeviceName Splitted %+v", realDeviceNameSplitted)
	log.Printf("[DEBUG] RealDeviceName Splitted Length %d", len(realDeviceNameSplitted))
	log.Printf("[DEBUG] RealDeviceName Splitted [0] %s", realDeviceNameSplitted[0])
	log.Printf("[DEBUG] RealDeviceName Splitted [1] %s", realDeviceNameSplitted[1])

	realDeviceNameStr := realDeviceNameSplitted[0]
	if realDeviceNameStr == "" {
		realDeviceNameStr = realDeviceNameSplitted[1]
	}

	deviceMount := fmt.Sprintf("/dev/%s", strings.Replace(realDeviceNameStr, "\n", "", -1))

	log.Printf("[DEBUG] s.MountPartition  = %s", s.MountPartition)
	log.Printf("[DEBUG ] DeviceMount: %s", deviceMount)

	if virtualizationType == "hvm" && s.MountPartition != "0" {
		deviceMount = fmt.Sprintf("%s%s", deviceMount, s.MountPartition)
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
	cmd = ShellCommand(mountCommand)
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
