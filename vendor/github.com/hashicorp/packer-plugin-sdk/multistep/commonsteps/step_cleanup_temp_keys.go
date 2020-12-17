package commonsteps

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type StepCleanupTempKeys struct {
	Comm *communicator.Config
}

func (s *StepCleanupTempKeys) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
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

	comm := state.Get("communicator").(packersdk.Communicator)
	ui := state.Get("ui").(packersdk.Ui)

	cmd := new(packersdk.RemoteCmd)

	ui.Say("Trying to remove ephemeral keys from authorized_keys files")

	// Per the OpenSSH manual (https://man.openbsd.org/sshd.8), a typical
	// line in the 'authorized_keys' file contains several fields that
	// are delimited by spaces. Here is an (abbreviated) example of a line:
	// 	ssh-rsa AAAAB3Nza...LiPk== user@example.net
	//
	// In the above example, 'ssh-rsa' is the key pair type,
	// 'AAAAB3Nza...LiPk==' is the base64 encoded public key,
	// and 'user@example.net' is a comment (in this case, describing
	// who the key belongs to).
	//
	// In the following 'sed' calls, the comment field will be equal to
	// the value of communicator.Config.SSHTemporaryKeyPairName.
	// We can remove an authorized public key using 'sed' by looking
	// for a line ending in ' packer-key-pair-comment' (note the
	// leading space).
	//
	// TODO: Why create a backup file if you are going to remove it?
	cmd.Command = fmt.Sprintf("sed -i.bak '/ %s$/d' ~/.ssh/authorized_keys; rm ~/.ssh/authorized_keys.bak", s.Comm.SSHTemporaryKeyPairName)
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		log.Printf("Error cleaning up ~/.ssh/authorized_keys; please clean up keys manually: %s", err)
	}
	cmd = new(packersdk.RemoteCmd)
	cmd.Command = fmt.Sprintf("sudo sed -i.bak '/ %s$/d' /root/.ssh/authorized_keys; sudo rm /root/.ssh/authorized_keys.bak", s.Comm.SSHTemporaryKeyPairName)
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		log.Printf("Error cleaning up /root/.ssh/authorized_keys; please clean up keys manually: %s", err)
	}

	return multistep.ActionContinue
}

func (s *StepCleanupTempKeys) Cleanup(state multistep.StateBag) {
}
