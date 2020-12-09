package communicator

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepDumpSSHKey is a multistep Step implementation that writes the ssh
// keypair somewhere.
type StepDumpSSHKey struct {
	Path string
	SSH  *SSH
}

func (s *StepDumpSSHKey) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	ui.Message(fmt.Sprintf("Saving key for debug purposes: %s", s.Path))

	err := ioutil.WriteFile(s.Path, s.SSH.SSHPrivateKey, 0700)
	if err != nil {
		state.Put("error", fmt.Errorf("Error saving debug key: %s", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepDumpSSHKey) Cleanup(state multistep.StateBag) {}
