package googlecompute

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// StepUpdateGsutil represents a Packer build step that updates the gsutil
// utility to the latest version available.
type StepUpdateGsutil int

// Run executes the Packer build step that updates the gsutil utility to the
// latest version available.
//
// This step is required to prevent the image creation process from hanging;
// the image creation process utilizes the gcimagebundle cli tool which will
// prompt to update gsutil if a newer version is available.
func (s *StepUpdateGsutil) Run(state multistep.StateBag) multistep.StepAction {
	comm := state.Get("communicator").(packer.Communicator)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	sudoPrefix := ""

	if config.SSHUsername != "root" {
		sudoPrefix = "sudo "
	}

	gsutilUpdateCmd := "/usr/local/bin/gsutil update -n -f"
	cmd := new(packer.RemoteCmd)
	cmd.Command = fmt.Sprintf("%s%s", sudoPrefix, gsutilUpdateCmd)

	ui.Say("Updating gsutil...")
	err := cmd.StartWithUi(comm, ui)
	if err == nil && cmd.ExitStatus != 0 {
		err = fmt.Errorf(
			"gsutil update exited with non-zero exit status: %d", cmd.ExitStatus)
	}
	if err != nil {
		err := fmt.Errorf("Error updating gsutil: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

// Cleanup.
func (s *StepUpdateGsutil) Cleanup(state multistep.StateBag) {}
