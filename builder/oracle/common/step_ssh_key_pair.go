package common

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"runtime"

	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"golang.org/x/crypto/ssh"
)

type StepKeyPair struct {
	Debug        bool
	Comm         *communicator.Config
	DebugKeyPath string
}

func (s *StepKeyPair) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	if s.Comm.SSHPrivateKeyFile != "" {
		ui.Say("Using existing SSH private key")
		privateKeyBytes, err := s.Comm.ReadSSHPrivateKeyFile()
		if err != nil {
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}

		key, err := ssh.ParsePrivateKey(privateKeyBytes)
		if err != nil {
			err = fmt.Errorf("Error parsing 'ssh_private_key_file': %s", err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}

		s.Comm.SSHPublicKey = ssh.MarshalAuthorizedKey(key.PublicKey())
		s.Comm.SSHPrivateKey = privateKeyBytes

		return multistep.ActionContinue
	}

	ui.Say("Creating temporary ssh key for instance...")

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		err = fmt.Errorf("Error creating temporary SSH key: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// ASN.1 DER encoded form
	privDer := x509.MarshalPKCS1PrivateKey(priv)
	privBlk := pem.Block{Type: "RSA PRIVATE KEY", Headers: nil, Bytes: privDer}

	// Set the private key in the statebag for later
	state.Put("privateKey", string(pem.EncodeToMemory(&privBlk)))

	// Marshal the public key into SSH compatible format
	pub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		err = fmt.Errorf("Error marshaling temporary SSH public key: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	s.Comm.SSHPublicKey = ssh.MarshalAuthorizedKey(pub)

	// If we're in debug mode, output the private key to the working
	// directory.
	if s.Debug {
		ui.Message(fmt.Sprintf("Saving key for debug purposes: %s", s.DebugKeyPath))
		f, err := os.Create(s.DebugKeyPath)
		if err != nil {
			err = fmt.Errorf("Error saving debug key: %s", err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
		defer f.Close()

		// Write the key out
		if _, err := f.Write(pem.EncodeToMemory(&privBlk)); err != nil {
			err = fmt.Errorf("Error saving debug key: %s", err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}

		// Chmod it so that it is SSH ready
		if runtime.GOOS != "windows" {
			if err := f.Chmod(0600); err != nil {
				err = fmt.Errorf("Error setting permissions of debug key: %s", err)
				ui.Error(err.Error())
				state.Put("error", err)
				return multistep.ActionHalt
			}
		}
	}

	return multistep.ActionContinue
}

func (s *StepKeyPair) Cleanup(state multistep.StateBag) {
	// Nothing to do
}
