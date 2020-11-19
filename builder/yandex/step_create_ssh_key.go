package yandex

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"golang.org/x/crypto/ssh"
)

type StepCreateSSHKey struct {
	Debug        bool
	DebugKeyPath string
}

func (s *StepCreateSSHKey) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	config := state.Get("config").(*Config)

	if config.Communicator.SSHPrivateKeyFile != "" {
		ui.Say("Using existing SSH private key")
		privateKeyBytes, err := config.Communicator.ReadSSHPrivateKeyFile()
		if err != nil {
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

		config.Communicator.SSHPublicKey = ssh.MarshalAuthorizedKey(key.PublicKey())
		config.Communicator.SSHPrivateKey = privateKeyBytes

		return multistep.ActionContinue
	}

	ui.Say("Creating temporary ssh key for instance...")

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return stepHaltWithError(state, fmt.Errorf("Error generating temporary SSH key: %s", err))
	}

	// ASN.1 DER encoded form
	privDer := x509.MarshalPKCS1PrivateKey(priv)
	privBlk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDer,
	}

	// Marshal the public key into SSH compatible format
	pub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		err = fmt.Errorf("Error creating public ssh key: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	pubSSHFormat := string(ssh.MarshalAuthorizedKey(pub))

	hashMD5 := ssh.FingerprintLegacyMD5(pub)
	hashSHA256 := ssh.FingerprintSHA256(pub)

	log.Printf("[INFO] md5 hash of ssh pub key: %s", hashMD5)
	log.Printf("[INFO] sha256 hash of ssh pub key: %s", hashSHA256)

	// Remember some state for the future
	state.Put("ssh_key_public", pubSSHFormat)

	// Set the private key in the config for later
	config.Communicator.SSHPrivateKey = pem.EncodeToMemory(&privBlk)
	config.Communicator.SSHPublicKey = ssh.MarshalAuthorizedKey(pub)

	// If we're in debug mode, output the private key to the working directory.
	if s.Debug {
		ui.Message(fmt.Sprintf("Saving key for debug purposes: %s", s.DebugKeyPath))
		err := ioutil.WriteFile(s.DebugKeyPath, config.Communicator.SSHPrivateKey, 0600)
		if err != nil {
			return stepHaltWithError(state, fmt.Errorf("Error saving debug key: %s", err))
		}
	}

	return multistep.ActionContinue
}

func (s *StepCreateSSHKey) Cleanup(state multistep.StateBag) {
}
