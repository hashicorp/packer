package common

import (
	"bytes"
	"fmt"
	"log"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// StepUploadVersion is a step that uploads a file containing the version of
// Parallels Desktop, which can be useful for various provisioning reasons.
//
// Uses:
//   communicator packer.Communicator
//   driver Driver
//   ui packer.Ui
type StepUploadVersion struct {
	Path string
}

// Run uploads a file containing the version of Parallels Desktop.
func (s *StepUploadVersion) Run(state multistep.StateBag) multistep.StepAction {
	comm := state.Get("communicator").(packer.Communicator)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	if s.Path == "" {
		log.Println("ParallelsVersionFile is empty. Not uploading.")
		return multistep.ActionContinue
	}

	version, err := driver.Version()
	if err != nil {
		state.Put("error", fmt.Errorf("Error reading version for metadata upload: %s", err))
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Uploading Parallels version info (%s)", version))
	var data bytes.Buffer
	data.WriteString(version)
	if err := comm.Upload(s.Path, &data, nil); err != nil {
		state.Put("error", fmt.Errorf("Error uploading Parallels version: %s", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

// Cleanup does nothing.
func (s *StepUploadVersion) Cleanup(state multistep.StateBag) {}
