package digitalocean

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/digitalocean/godo"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/mitchellh/packer/packer"
	"golang.org/x/crypto/ssh"
)

type stepCreateSSHKey struct {
	Debug        bool
	DebugKeyPath string

	keyId int
}

func (s *stepCreateSSHKey) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*godo.Client)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating temporary ssh key for droplet...")

	priv, err := rsa.GenerateKey(rand.Reader, 2014)

	// ASN.1 DER encoded form
	priv_der := x509.MarshalPKCS1PrivateKey(priv)
	priv_blk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   priv_der,
	}

	// Set the private key in the statebag for later
	state.Put("privateKey", string(pem.EncodeToMemory(&priv_blk)))

	// Marshal the public key into SSH compatible format
	// TODO properly handle the public key error
	pub, _ := ssh.NewPublicKey(&priv.PublicKey)
	pub_sshformat := string(ssh.MarshalAuthorizedKey(pub))

	// The name of the public key on DO
	name := fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())

	// Create the key!
	key, _, err := client.Keys.Create(&godo.KeyCreateRequest{
		Name:      name,
		PublicKey: pub_sshformat,
	})
	if err != nil {
		err := fmt.Errorf("Error creating temporary SSH key: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// We use this to check cleanup
	s.keyId = key.ID

	log.Printf("temporary ssh key name: %s", name)

	// Remember some state for the future
	state.Put("ssh_key_id", key.ID)

	// If we're in debug mode, output the private key to the working directory.
	if s.Debug {
		ui.Message(fmt.Sprintf("Saving key for debug purposes: %s", s.DebugKeyPath))
		f, err := os.Create(s.DebugKeyPath)
		if err != nil {
			state.Put("error", fmt.Errorf("Error saving debug key: %s", err))
			return multistep.ActionHalt
		}
		defer f.Close()

		// Write the key out
		if _, err := f.Write(pem.EncodeToMemory(&priv_blk)); err != nil {
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

func (s *stepCreateSSHKey) Cleanup(state multistep.StateBag) {
	// If no key name is set, then we never created it, so just return
	if s.keyId == 0 {
		return
	}

	client := state.Get("client").(*godo.Client)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting temporary ssh key...")
	_, err := client.Keys.DeleteByID(s.keyId)
	if err != nil {
		log.Printf("Error cleaning up ssh key: %s", err)
		ui.Error(fmt.Sprintf(
			"Error cleaning up ssh key. Please delete the key manually: %s", err))
	}
}
