package common

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// This step uploads a file containing the VirtualBox version, which
// can be useful for various provisioning reasons.
type StepUploadVersion struct {
	Path string
}

func (s *StepUploadVersion) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	comm := state.Get("communicator").(packer.Communicator)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	if s.Path == "" {
		log.Println("VBoxVersionFile is empty. Not uploading.")
		return multistep.ActionContinue
	}

	version, err := driver.Version()
	if err != nil {
		state.Put("error", fmt.Errorf("Error reading version for metadata upload: %s", err))
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Uploading VirtualBox version info (%s)", version))
	var data bytes.Buffer
	data.WriteString(version)
	if err := comm.Upload(s.Path, &data, nil); err != nil {
		state.Put("error", fmt.Errorf("Error uploading VirtualBox version: %s", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepUploadVersion) Cleanup(state multistep.StateBag) {}
