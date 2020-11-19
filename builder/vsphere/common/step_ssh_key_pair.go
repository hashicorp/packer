package common

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/communicator/ssh"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/uuid"
)

// StepSshKeyPair executes the business logic for setting the SSH key pair in
// the specified communicator.Config.
type StepSshKeyPair struct {
	Debug        bool
	DebugKeyPath string
	Comm         *communicator.Config
}

func (s *StepSshKeyPair) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if s.Comm.Type != "ssh" || s.Comm.SSHPassword != "" {
		return multistep.ActionContinue
	}

	ui := state.Get("ui").(packersdk.Ui)

	comment := fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID())
	if s.Comm.SSHPrivateKeyFile != "" {
		ui.Say("Using existing SSH private key for the communicator...")
		privateKeyBytes, err := s.Comm.ReadSSHPrivateKeyFile()
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}

		kp, err := ssh.KeyPairFromPrivateKey(ssh.FromPrivateKeyConfig{
			RawPrivateKeyPemBlock: privateKeyBytes,
			Comment:               comment,
		})
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}

		s.Comm.SSHPrivateKey = privateKeyBytes
		s.Comm.SSHKeyPairName = kp.Comment
		s.Comm.SSHTemporaryKeyPairName = kp.Comment
		s.Comm.SSHPublicKey = kp.PublicKeyAuthorizedKeysLine

		return multistep.ActionContinue
	}

	if s.Comm.SSHAgentAuth {
		ui.Say("Using local SSH Agent to authenticate connections for the communicator...")
		return multistep.ActionContinue
	}

	ui.Say("Creating ephemeral key pair for SSH communicator...")

	if s.Comm.SSHTemporaryKeyPairName != "" {
		comment = s.Comm.SSHTemporaryKeyPairName
	}

	kp, err := ssh.NewKeyPair(ssh.CreateKeyPairConfig{
		Comment: comment,
		Type:    ssh.Rsa,
	})
	if err != nil {
		state.Put("error", fmt.Errorf("Error creating temporary keypair: %s", err))
		return multistep.ActionHalt
	}

	s.Comm.SSHKeyPairName = kp.Comment
	s.Comm.SSHTemporaryKeyPairName = kp.Comment
	s.Comm.SSHPrivateKey = kp.PrivateKeyPemBlock
	s.Comm.SSHPublicKey = kp.PublicKeyAuthorizedKeysLine
	s.Comm.SSHClearAuthorizedKeys = true

	vm := state.Get("vm").(*driver.VirtualMachineDriver)
	err = vm.AddPublicKeys(ctx, string(s.Comm.SSHPublicKey))
	if err != nil {
		state.Put("error", fmt.Errorf("error saving temporary keypair in the vm: %s", err))
		return multistep.ActionHalt
	}

	ui.Say("Created ephemeral SSH key pair for communicator")

	// If we're in debug mode, output the private key to the working
	// directory.
	if s.Debug {
		ui.Message(fmt.Sprintf("Saving communicator private key for debug purposes: %s", s.DebugKeyPath))
		// Write the key out
		if err := ioutil.WriteFile(s.DebugKeyPath, kp.PrivateKeyPemBlock, 0600); err != nil {
			state.Put("error", fmt.Errorf("Error saving debug key: %s", err))
			return multistep.ActionHalt
		}
	}

	return multistep.ActionContinue
}

func (s *StepSshKeyPair) Cleanup(state multistep.StateBag) {
	if s.Debug {
		if err := os.Remove(s.DebugKeyPath); err != nil {
			ui := state.Get("ui").(packersdk.Ui)
			ui.Error(fmt.Sprintf(
				"Error removing debug key '%s': %s", s.DebugKeyPath, err))
		}
	}
}
