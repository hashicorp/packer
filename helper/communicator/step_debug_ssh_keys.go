package communicator

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepDumpSSHKey is a multistep Step implementation that writes the ssh
// keypair somewhere.
type StepDumpSSHKey struct {
	Path string
}

func (s *StepDumpSSHKey) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	// Put communicator config into state so we can pass it to provisioners
	// for specialized interpolation later
	comm := state.Get("communicator_config").(Config)

	ui.Message(fmt.Sprintf("Saving key for debug purposes: %s", s.Path))

	err := ioutil.WriteFile(s.Path, comm.SSHPrivateKey, 0700)
	if err != nil {
		state.Put("error", fmt.Errorf("Error saving debug key: %s", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepDumpSSHKey) Cleanup(state multistep.StateBag) {}
