package communicator

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/communicator/sshkey"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepSSHKeyGen is a Packer build step that generates SSH key pairs.
type StepSSHKeyGen struct {
	CommConf *Config
	SSHTemporaryKeyPair
}

// Run executes the Packer build step that generates SSH key pairs.
// The key pairs are added to the ssh config
func (s *StepSSHKeyGen) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	comm := s.CommConf

	if comm.SSHPrivateKeyFile != "" {
		ui.Say("Using existing SSH private key")
		privateKeyBytes, err := comm.ReadSSHPrivateKeyFile()
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}

		comm.SSHPrivateKey = privateKeyBytes
		comm.SSHPublicKey = nil

		return multistep.ActionContinue
	}

	algorithm := s.SSHTemporaryKeyPair.SSHTemporaryKeyPairType
	if algorithm == "" {
		algorithm = sshkey.RSA.String()
	}
	a, err := sshkey.AlgorithmString(algorithm)
	if err != nil {
		err := fmt.Errorf("%w: possible algorithm types are `dsa` | `ecdsa` | `ed25519` | `rsa` ( the default )", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Creating temporary %s SSH key for instance...", a.String()))
	pair, err := sshkey.GeneratePair(a, nil, s.SSHTemporaryKeyPairBits)
	if err != nil {
		err := fmt.Errorf("Error creating temporary ssh key: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	comm.SSHPrivateKey = pair.Private
	comm.SSHPublicKey = pair.Public

	return multistep.ActionContinue
}

// Nothing to clean up. SSH keys are associated with a single GCE instance.
func (s *StepSSHKeyGen) Cleanup(state multistep.StateBag) {}
