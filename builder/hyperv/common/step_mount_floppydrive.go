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

const (
	FloppyFileName = "assets.vfd"
)

type StepMountFloppydrive struct {
	Generation uint
	floppyPath string
}

func (s *StepMountFloppydrive) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if s.Generation > 1 {
		return multistep.ActionContinue
	}

	driver := state.Get("driver").(Driver)

	// Determine if we even have a floppy disk to attach
	var floppyPath string
	if floppyPathRaw, ok := state.GetOk("floppy_path"); ok {
		floppyPath = floppyPathRaw.(string)
	} else {
		log.Println("No floppy disk, not attaching.")
		return multistep.ActionContinue
	}

	// Hyper-V is really dumb and can't figure out the format of the file
	// without an extension, so we need to add the "vfd" extension to the
	// floppy.
	floppyPath, err := s.copyFloppy(floppyPath)
	if err != nil {
		state.Put("error", fmt.Errorf("Error preparing floppy: %s", err))
		return multistep.ActionHalt
	}

	ui := state.Get("ui").(packersdk.Ui)
	vmName := state.Get("vmName").(string)

	ui.Say("Mounting floppy drive...")

	err = driver.MountFloppyDrive(vmName, floppyPath)
	if err != nil {
		state.Put("error", fmt.Errorf("Error mounting floppy drive: %s", err))
		return multistep.ActionHalt
	}

	// Track the path so that we can unregister it from Hyper-V later
	s.floppyPath = floppyPath

	return multistep.ActionContinue
}

func (s *StepMountFloppydrive) Cleanup(state multistep.StateBag) {
	if s.Generation > 1 {
		return
	}
	driver := state.Get("driver").(Driver)
	if s.floppyPath == "" {
		return
	}

	errorMsg := "Error unmounting floppy drive: %s"

	vmName := state.Get("vmName").(string)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Cleanup floppy drive...")

	err := driver.UnmountFloppyDrive(vmName)
	if err != nil {
		log.Print(fmt.Sprintf(errorMsg, err))
	}

	err = os.Remove(s.floppyPath)

	if err != nil {
		log.Print(fmt.Sprintf(errorMsg, err))
	}
}

func (s *StepMountFloppydrive) copyFloppy(path string) (string, error) {
	tempdir, err := tmp.Dir("hyperv")
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
