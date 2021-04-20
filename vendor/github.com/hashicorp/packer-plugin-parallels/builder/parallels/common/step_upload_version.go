package common

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepUploadVersion is a step that uploads a file containing the version of
// Parallels Desktop, which can be useful for various provisioning reasons.
//
// Uses:
//   communicator packersdk.Communicator
//   driver Driver
//   ui packersdk.Ui
type StepUploadVersion struct {
	Path string
}

// Run uploads a file containing the version of Parallels Desktop.
func (s *StepUploadVersion) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	comm := state.Get("communicator").(packersdk.Communicator)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

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
