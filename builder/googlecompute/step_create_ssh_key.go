package googlecompute

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"golang.org/x/crypto/ssh"
)

// StepCreateSSHKey represents a Packer build step that generates SSH key pairs.
type StepCreateSSHKey struct {
	Debug          bool
	DebugKeyPath   string
	PrivateKeyFile string
}

// Run executes the Packer build step that generates SSH key pairs.
// The key pairs are added to the ssh config
func (s *StepCreateSSHKey) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)

	if s.PrivateKeyFile != "" {
		ui.Say("Using existing SSH private key")
		privateKeyBytes, err := ioutil.ReadFile(s.PrivateKeyFile)
		if err != nil {
			state.Put("error", fmt.Errorf(
				"Error loading configured private key file: %s", err))
			return multistep.ActionHalt
		}

		config.Comm.SSHPrivateKey = privateKeyBytes
		config.Comm.SSHPublicKey = nil

		return multistep.ActionContinue
	}

	ui.Say("Creating temporary SSH key for instance...")
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		err := fmt.Errorf("Error creating temporary ssh key: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	priv_blk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   x509.MarshalPKCS1PrivateKey(priv),
	}

	pub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		err := fmt.Errorf("Error creating temporary ssh key: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	config.Comm.SSHPrivateKey = pem.EncodeToMemory(&priv_blk)
	config.Comm.SSHPublicKey = ssh.MarshalAuthorizedKey(pub)

	if s.Debug {
		ui.Message(fmt.Sprintf("Saving key for debug purposes: %s", s.DebugKeyPath))
		f, err := os.Create(s.DebugKeyPath)
		if err != nil {
			state.Put("error", fmt.Errorf("Error saving debug key: %s", err))
			return multistep.ActionHalt
		}

		// Write out the key
		err = pem.Encode(f, &priv_blk)
		f.Close()
		if err != nil {
			state.Put("error", fmt.Errorf("Error saving debug key: %s", err))
			return multistep.ActionHalt
		}
	}
	return multistep.ActionContinue
}

// Nothing to clean up. SSH keys are associated with a single GCE instance.
func (s *StepCreateSSHKey) Cleanup(state multistep.StateBag) {}
