package common

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/tmp"
)

// This step attaches the ISO to the virtual machine.
//
// Uses:
//   driver Driver
//   ui packersdk.Ui
//   vmName string
//
// Produces:
type StepAttachFloppy struct {
	floppyPath string
}

func (s *StepAttachFloppy) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// Determine if we even have a floppy disk to attach
	var floppyPath string
	if floppyPathRaw, ok := state.GetOk("floppy_path"); ok {
		floppyPath = floppyPathRaw.(string)
	} else {
		log.Println("No floppy disk, not attaching.")
		return multistep.ActionContinue
	}

	// VirtualBox is really dumb and can't figure out the format of the file
	// without an extension, so we need to add the "vfd" extension to the
	// floppy.
	floppyPath, err := s.copyFloppy(floppyPath)
	if err != nil {
		state.Put("error", fmt.Errorf("Error preparing floppy: %s", err))
		return multistep.ActionHalt
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)
	vmName := state.Get("vmName").(string)

	ui.Say("Deleting any current floppy disk...")
	if err := driver.RemoveFloppyControllers(vmName); err != nil {
		state.Put("error", fmt.Errorf("Error deleting existing floppy controllers: %s", err))
		return multistep.ActionHalt
	}

	ui.Say("Attaching floppy disk...")

	// Create the floppy disk controller
	command := []string{
		"storagectl", vmName,
		"--name", "Floppy Controller",
		"--add", "floppy",
	}
	if err := driver.VBoxManage(command...); err != nil {
		state.Put("error", fmt.Errorf("Error creating floppy controller: %s", err))
		return multistep.ActionHalt
	}

	// Attach the floppy to the controller
	command = []string{
		"storageattach", vmName,
		"--storagectl", "Floppy Controller",
		"--port", "0",
		"--device", "0",
		"--type", "fdd",
		"--medium", floppyPath,
	}
	if err := driver.VBoxManage(command...); err != nil {
		state.Put("error", fmt.Errorf("Error attaching floppy: %s", err))
		return multistep.ActionHalt
	}

	// Track the path so that we can unregister it from VirtualBox later
	s.floppyPath = floppyPath

	return multistep.ActionContinue
}

func (s *StepAttachFloppy) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Cleaning up floppy disk...")
	if s.floppyPath == "" {
		return
	}

	// Delete the floppy disk
	defer os.Remove(s.floppyPath)

	driver := state.Get("driver").(Driver)
	vmName := state.Get("vmName").(string)

	command := []string{
		"storageattach", vmName,
		"--storagectl", "Floppy Controller",
		"--port", "0",
		"--device", "0",
		"--medium", "none",
	}

	if err := driver.VBoxManage(command...); err != nil {
		ui.Error(fmt.Sprintf("Error unregistering floppy: %s. "+
			"Not considering this a critical failure; build will continue.", err))
	}
}

func (s *StepAttachFloppy) copyFloppy(path string) (string, error) {
	tempdir, err := tmp.Dir("virtualbox")
	if err != nil {
		return "", err
	}

	floppyPath := filepath.Join(tempdir, "floppy.vfd")
	f, err := os.Create(floppyPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	sourceF, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer sourceF.Close()

	log.Printf("Copying floppy to temp location: %s", floppyPath)
	if _, err := io.Copy(f, sourceF); err != nil {
		return "", err
	}

	return floppyPath, nil
}
