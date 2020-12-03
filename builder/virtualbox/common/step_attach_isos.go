package common

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// This step attaches the boot ISO, cd_files iso, and guest additions to the
// virtual machine, if present.
type StepAttachISOs struct {
	AttachBootISO           bool
	ISOInterface            string
	GuestAdditionsMode      string
	GuestAdditionsInterface string
	diskUnmountCommands     map[string][]string
}

func (s *StepAttachISOs) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// Check whether there is anything to attach
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Mounting ISOs...")
	diskMountMap := map[string]string{}
	s.diskUnmountCommands = map[string][]string{}
	// Track the bootable iso (only used in virtualbox-iso builder. )
	if s.AttachBootISO {
		isoPath := state.Get("iso_path").(string)
		diskMountMap["boot_iso"] = isoPath
	}

	// Determine if we even have a cd_files disk to attach
	if cdPathRaw, ok := state.GetOk("cd_path"); ok {
		cdFilesPath := cdPathRaw.(string)
		diskMountMap["cd_files"] = cdFilesPath
	}

	// Determine if we have guest additions to attach
	if s.GuestAdditionsMode != GuestAdditionsModeAttach {
		log.Println("Not attaching guest additions since we're uploading.")
	} else {
		// Get the guest additions path since we're doing it
		guestAdditionsPath := state.Get("guest_additions_path").(string)
		diskMountMap["guest_additions"] = guestAdditionsPath
	}

	if len(diskMountMap) == 0 {
		ui.Message("No ISOs to mount; continuing...")
		return multistep.ActionContinue
	}

	driver := state.Get("driver").(Driver)
	vmName := state.Get("vmName").(string)

	for diskCategory, isoPath := range diskMountMap {
		// If it's a symlink, resolve it to its target.
		resolvedIsoPath, err := filepath.EvalSymlinks(isoPath)
		if err != nil {
			err := fmt.Errorf("Error resolving symlink for ISO: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		isoPath = resolvedIsoPath

		// We have three different potential isos we can attach, so let's
		// assign each one its own spot so they don't conflict.
		var controllerName, device, port string
		switch diskCategory {
		case "boot_iso":
			// figure out controller path
			controllerName = "IDE Controller"
			port = "0"
			device = "1"
			if s.ISOInterface == "sata" {
				controllerName = "SATA Controller"
				port = "1"
				device = "0"
			}
			ui.Message("Mounting boot ISO...")
		case "guest_additions":
			controllerName = "IDE Controller"
			port = "1"
			device = "0"
			if s.GuestAdditionsInterface == "sata" {
				controllerName = "SATA Controller"
				port = "2"
				device = "0"
			}
			ui.Message("Mounting guest additions ISO...")
		case "cd_files":
			controllerName = "IDE Controller"
			port = "1"
			device = "1"
			if s.ISOInterface == "sata" {
				controllerName = "SATA Controller"
				port = "3"
				device = "0"
			}
			ui.Message("Mounting cd_files ISO...")
		}

		// Attach the disk to the controller
		command := []string{
			"storageattach", vmName,
			"--storagectl", controllerName,
			"--port", port,
			"--device", device,
			"--type", "dvddrive",
			"--medium", isoPath,
		}
		if err := driver.VBoxManage(command...); err != nil {
			err := fmt.Errorf("Error attaching ISO: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Track the disks we've mounted so we can remove them without having
		// to re-derive what was mounted where
		unmountCommand := []string{
			"storageattach", vmName,
			"--storagectl", controllerName,
			"--port", port,
			"--device", device,
			"--type", "dvddrive",
			"--medium", "none",
		}

		s.diskUnmountCommands[diskCategory] = unmountCommand
	}

	state.Put("disk_unmount_commands", s.diskUnmountCommands)

	return multistep.ActionContinue
}

func (s *StepAttachISOs) Cleanup(state multistep.StateBag) {
	if len(s.diskUnmountCommands) == 0 {
		return
	}

	driver := state.Get("driver").(Driver)
	_, ok := state.GetOk("detached_isos")

	if !ok {
		for _, command := range s.diskUnmountCommands {
			err := driver.VBoxManage(command...)
			if err != nil {
				log.Printf("error detaching iso: %s", err)
			}
		}
	}
}
