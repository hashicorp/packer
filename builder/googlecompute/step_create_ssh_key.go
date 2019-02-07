package googlecompute

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"golang.org/x/crypto/ed25519"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"golang.org/x/crypto/ssh"
)

// StepCreateSSHKey represents a Packer build step that generates SSH key pairs.
type StepCreateSSHKey struct {
	Debug        bool
	DebugKeyPath string
}

func createTransientPublicKey(pemBytes []byte) (ssh.PublicKey, error) {
	privateKey, err := ssh.ParseRawPrivateKey(pemBytes)
	if err != nil {
		return nil, err
	}

	var publicKey interface{}
	switch key := privateKey.(type) {
	case *rsa.PrivateKey:
		publicKey = key.Public()
	case *ecdsa.PrivateKey:
		publicKey = key.Public()
	case *ed25519.PrivateKey:
		publicKey = key.Public()
	default:
		return nil, fmt.Errorf("ssh: unsupported key type")
	}

	pub, err := ssh.NewPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("unsupported key type detected: %s", err)
	}

	return pub, nil
}

// Run executes the Packer build step that generates SSH key pairs.
// The key pairs are added to the ssh config
func (s *StepCreateSSHKey) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)

	if config.Comm.SSHPrivateKeyFile != "" {
		ui.Say("Using existing SSH private key")
		privateKeyBytes, err := config.Comm.ReadSSHPrivateKeyFile()
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}

		pub, err := createTransientPublicKey(privateKeyBytes)
		if err != nil {
			err := fmt.Errorf("Error creating temporary ssh key: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		config.Comm.SSHPrivateKey = privateKeyBytes
		config.Comm.SSHPublicKey = ssh.MarshalAuthorizedKey(pub)

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
