package qemu

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// This step attaches the ISO to the virtual machine.
//
// Uses:
//
// Produces:
type stepCopyFloppy struct {
	floppyPath string
}

func (s *stepCopyFloppy) Run(state multistep.StateBag) multistep.StepAction {
	// Determine if we even have a floppy disk to attach
	var floppyPath string
	if floppyPathRaw, ok := state.GetOk("floppy_path"); ok {
		floppyPath = floppyPathRaw.(string)
	} else {
		log.Println("No floppy disk, not attaching.")
		return multistep.ActionContinue
	}

	// copy the floppy for exclusive use during the vm creation
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Copying floppy disk for exclusive use...")
	floppyPath, err := s.copyFloppy(floppyPath)
	if err != nil {
		state.Put("error", fmt.Errorf("Error preparing floppy: %s", err))
		return multistep.ActionHalt
	}

	// Track the path so that we can remove it later
	s.floppyPath = floppyPath

	return multistep.ActionContinue
}

func (s *stepCopyFloppy) Cleanup(state multistep.StateBag) {
	if s.floppyPath == "" {
		return
	}

	// Delete the floppy disk
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Removing floppy disk previously copied...")
	defer os.Remove(s.floppyPath)
}

func (s *stepCopyFloppy) copyFloppy(path string) (string, error) {
	tempdir, err := ioutil.TempDir("", "packer")
	if err != nil {
		return "", err
	}

	floppyPath := filepath.Join(tempdir, "floppy.img")
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
