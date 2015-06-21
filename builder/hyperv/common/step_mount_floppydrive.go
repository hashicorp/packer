// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"os"
	"strings"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/powershell"
	"github.com/mitchellh/packer/powershell/hyperv"
	"log"
	"io"
	"io/ioutil"
	"path/filepath"
)


const(
	FloppyFileName = "assets.vfd"
)




type StepSetUnattendedProductKey struct {
	Files []string
	ProductKey string
}

func (s *StepSetUnattendedProductKey) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	if s.ProductKey == "" {
		ui.Say("No product key specified...")
		return multistep.ActionContinue 
	}

	index := -1
	for i, value := range s.Files {
    	if s.caseInsensitiveContains(value, "Autounattend.xml") {
    		index = i
    		break
    	}
	}

	ui.Say("Setting product key in Autounattend.xml...")
	copyOfAutounattend, err := s.copyAutounattend(s.Files[index])
	if err != nil {
		state.Put("error", fmt.Errorf("Error copying Autounattend.xml: %s", err))
		return multistep.ActionHalt
	}

	powershell.SetUnattendedProductKey(copyOfAutounattend, s.ProductKey)
	s.Files[index] = copyOfAutounattend
	return multistep.ActionContinue
}


func (s *StepSetUnattendedProductKey) caseInsensitiveContains(str, substr string) bool {
    str, substr = strings.ToUpper(str), strings.ToUpper(substr)
    return strings.Contains(str, substr)
}

func (s *StepSetUnattendedProductKey) copyAutounattend(path string) (string, error) {
	tempdir, err := ioutil.TempDir("", "packer")
	if err != nil {
		return "", err
	}

	autounattend := filepath.Join(tempdir, "Autounattend.xml")
	f, err := os.Create(autounattend)
	if err != nil {
		return "", err
	}
	defer f.Close()

	sourceF, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer sourceF.Close()

	log.Printf("Copying %s to temp location: %s", path, autounattend)
	if _, err := io.Copy(f, sourceF); err != nil {
		return "", err
	}

	return autounattend, nil
}


func (s *StepSetUnattendedProductKey) Cleanup(state multistep.StateBag) {
}



type StepMountFloppydrive struct {
	floppyPath string
}

func (s *StepMountFloppydrive) Run(state multistep.StateBag) multistep.StepAction {
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

	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	ui.Say("Mounting floppy drive...")

	err = hyperv.MountFloppyDrive(vmName, floppyPath)
	if err != nil {
		state.Put("error", fmt.Errorf("Error mounting floppy drive: %s", err))
		return multistep.ActionHalt
	}

	// Track the path so that we can unregister it from Hyper-V later
	s.floppyPath = floppyPath

	return multistep.ActionContinue}

func (s *StepMountFloppydrive) Cleanup(state multistep.StateBag) {
	if s.floppyPath == "" {
		return
	}

	errorMsg := "Error unmounting floppy drive: %s"

	vmName := state.Get("vmName").(string)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Unmounting floppy drive (cleanup)...")

	err := hyperv.UnmountFloppyDrive(vmName)
	if err != nil {
		ui.Error(fmt.Sprintf(errorMsg, err))
	}

	err = os.Remove(s.floppyPath)

	if err != nil {
		ui.Error(fmt.Sprintf(errorMsg, err))
	}
}

func (s *StepMountFloppydrive) copyFloppy(path string) (string, error) {
	tempdir, err := ioutil.TempDir("", "packer")
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
