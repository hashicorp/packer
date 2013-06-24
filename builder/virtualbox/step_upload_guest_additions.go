package virtualbox

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"os"
)

// This step uploads a file containing the VirtualBox version, which
// can be useful for various provisioning reasons.
type stepUploadGuestAdditions struct{}

func (s *stepUploadGuestAdditions) Run(state map[string]interface{}) multistep.StepAction {
	comm := state["communicator"].(packer.Communicator)
	config := state["config"].(*config)
	guestAdditionsPath := state["guest_additions_path"].(string)
	ui := state["ui"].(packer.Ui)

	f, err := os.Open(guestAdditionsPath)
	if err != nil {
		state["error"] = fmt.Errorf("Error opening guest additions ISO: %s", err)
		return multistep.ActionHalt
	}

	ui.Say("Upload VirtualBox guest additions ISO...")
	if err := comm.Upload(config.GuestAdditionsPath, f); err != nil {
		state["error"] = fmt.Errorf("Error uploading guest additions: %s", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepUploadGuestAdditions) Cleanup(state map[string]interface{}) {}
