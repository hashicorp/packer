package common

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepCleanupTempKeys struct {
	Comm *communicator.Config
}

func (s *StepCleanupTempKeys) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	// This step is mostly cosmetic; Packer deletes the ephemeral keys anyway
	// so there's no realistic situation where these keys can cause issues.
	// However, it's nice to clean up after yourself.

	if !s.Comm.SSHClearAuthorizedKeys {
		return multistep.ActionContinue
	}

	if s.Comm.Type != "ssh" {
		return multistep.ActionContinue
	}

	if s.Comm.SSHTemporaryKeyPairName == "" {
		return multistep.ActionContinue
	}

	comm := state.Get("communicator").(packer.Communicator)
	ui := state.Get("ui").(packer.Ui)

	cmd := new(packer.RemoteCmd)

	ui.Say("Trying to remove ephemeral keys from authorized_keys files")

	cmd.Command = fmt.Sprintf("sed -i.bak '/ %s$/d' ~/.ssh/authorized_keys; rm ~/.ssh/authorized_keys.bak", s.Comm.SSHTemporaryKeyPairName)
	if err := cmd.StartWithUi(comm, ui); err != nil {
		log.Printf("Error cleaning up ~/.ssh/authorized_keys; please clean up keys manually: %s", err)
	}
	cmd = new(packer.RemoteCmd)
	cmd.Command = fmt.Sprintf("sudo sed -i.bak '/ %s$/d' /root/.ssh/authorized_keys; sudo rm /root/.ssh/authorized_keys.bak", s.Comm.SSHTemporaryKeyPairName)

	if err := cmd.StartWithUi(comm, ui); err != nil {
		log.Printf("Error cleaning up /root/.ssh/authorized_keys; please clean up keys manually: %s", err)
	}

	return multistep.ActionContinue
}

func (s *StepCleanupTempKeys) Cleanup(state multistep.StateBag) {
}
