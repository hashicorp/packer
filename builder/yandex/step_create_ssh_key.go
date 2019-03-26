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

	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"golang.org/x/crypto/ssh"
)

type stepCreateSSHKey struct {
	Debug        bool
	DebugKeyPath string
}

func (s *stepCreateSSHKey) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(*Config)

	if c.Communicator.SSHPrivateKeyFile != "" {
		ui.Say("Using existing SSH private key")
		privateKeyBytes, err := c.Communicator.ReadSSHPrivateKeyFile()
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

		c.Communicator.SSHPublicKey = ssh.MarshalAuthorizedKey(key.PublicKey())
		c.Communicator.SSHPrivateKey = privateKeyBytes

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
	// TODO properly handle the public key error
	pub, _ := ssh.NewPublicKey(&priv.PublicKey)
	pubSSHFormat := string(ssh.MarshalAuthorizedKey(pub))

	// The name of the public key on DO
	name := fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())

	hashMd5 := ssh.FingerprintLegacyMD5(pub)
	hashSha256 := ssh.FingerprintSHA256(pub)

	log.Printf("[INFO] temporary ssh key name: %s", name)
	log.Printf("[INFO] md5 hash of ssh pub key: %s", hashMd5)
	log.Printf("[INFO] sha256 hash of ssh pub key: %s", hashSha256)

	// Remember some state for the future
	//state.Put("ssh_key_id", key.ID)
	state.Put("ssh_key_public", pubSSHFormat)
	state.Put("ssh_key_name", name)

	// Set the private key in the config for later
	c.Communicator.SSHPrivateKey = pem.EncodeToMemory(&privBlk)
	c.Communicator.SSHPublicKey = ssh.MarshalAuthorizedKey(pub)

	// If we're in debug mode, output the private key to the working directory.
	if s.Debug {
		ui.Message(fmt.Sprintf("Saving key for debug purposes: %s", s.DebugKeyPath))
		err := ioutil.WriteFile(s.DebugKeyPath, c.Communicator.SSHPrivateKey, 0600)
		if err != nil {
			return stepHaltWithError(state, fmt.Errorf("Error saving debug key: %s", err))
		}
	}

	return multistep.ActionContinue
}

func (s *stepCreateSSHKey) Cleanup(state multistep.StateBag) {
}
