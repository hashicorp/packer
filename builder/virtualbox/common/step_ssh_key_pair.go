package common

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/helper/ssh"
	"github.com/hashicorp/packer/packer"
)

// StepSshKeyPair executes the business logic for setting the SSH key pair in
// the specified communicator.Config.
type StepSshKeyPair struct {
	Debug        bool
	DebugKeyPath string
	Comm         *communicator.Config
}

func (s *StepSshKeyPair) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	if s.Comm.SSHPassword != "" {
		return multistep.ActionContinue
	}

	ui := state.Get("ui").(packer.Ui)

	if s.Comm.SSHPrivateKeyFile != "" {
		ui.Say("Using existing SSH private key for the communicator...")
		privateKeyBytes, err := s.Comm.ReadSSHPrivateKeyFile()
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}

		s.Comm.SSHPrivateKey = privateKeyBytes

		return multistep.ActionContinue
	}

	if s.Comm.SSHAgentAuth {
		ui.Say("Using local SSH Agent to authenticate connections for the communicator...")
		return multistep.ActionContinue
	}

	ui.Say("Creating ephemeral key pair for SSH communicator...")

	kp, err := ssh.NewKeyPairBuilder().
		SetName(fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID())).
		Build()
	if err != nil {
		state.Put("error", fmt.Errorf("Error creating temporary keypair: %s", err))
		return multistep.ActionHalt
	}

	s.Comm.SSHKeyPairName = kp.Name()
	s.Comm.SSHTemporaryKeyPairName = kp.Name()
	s.Comm.SSHPrivateKey = kp.PrivateKeyPemBlock()
	s.Comm.SSHPublicKey = kp.PublicKeyAuthorizedKeysFormat(ssh.UnixNewLine)
	s.Comm.SSHClearAuthorizedKeys = true

	ui.Say(fmt.Sprintf("Created ephemeral SSH key pair of type %s", kp.Description()))

	// If we're in debug mode, output the private key to the working
	// directory.
	if s.Debug {
		ui.Message(fmt.Sprintf("Saving communicator private key for debug purposes: %s", s.DebugKeyPath))
		f, err := os.Create(s.DebugKeyPath)
		if err != nil {
			state.Put("error", fmt.Errorf("Error saving debug key: %s", err))
			return multistep.ActionHalt
		}
		defer f.Close()

		// Write the key out
		if _, err := f.Write(kp.PrivateKeyPemBlock()); err != nil {
			state.Put("error", fmt.Errorf("Error saving debug key: %s", err))
			return multistep.ActionHalt
		}

		// Chmod it so that it is SSH ready
		if runtime.GOOS != "windows" {
			if err := f.Chmod(0600); err != nil {
				state.Put("error", fmt.Errorf("Error setting permissions of debug key: %s", err))
				return multistep.ActionHalt
			}
		}
	}

	return multistep.ActionContinue
}

func (s *StepSshKeyPair) Cleanup(state multistep.StateBag) {
	if s.Debug {
		if err := os.Remove(s.DebugKeyPath); err != nil {
			ui := state.Get("ui").(packer.Ui)
			ui.Error(fmt.Sprintf(
				"Error removing debug key '%s': %s", s.DebugKeyPath, err))
		}
	}
}
