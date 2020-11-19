package hcloud

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/uuid"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"golang.org/x/crypto/ssh"
)

type stepCreateSSHKey struct {
	Debug        bool
	DebugKeyPath string

	keyId int
}

func (s *stepCreateSSHKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("hcloudClient").(*hcloud.Client)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	ui.Say("Creating temporary ssh key for server...")

	priv, err := rsa.GenerateKey(rand.Reader, 2014)
	if err != nil {
		state.Put("error", fmt.Errorf("Error generating RSA key: %s", err))
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	// ASN.1 DER encoded form
	privDER := x509.MarshalPKCS1PrivateKey(priv)
	privBLK := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Set the private key in the config for later
	c.Comm.SSHPrivateKey = pem.EncodeToMemory(&privBLK)

	// Marshal the public key into SSH compatible format
	pub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		state.Put("error", fmt.Errorf("Error generating public key: %s", err))
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	pubSSHFormat := string(ssh.MarshalAuthorizedKey(pub))

	// The name of the public key on the Hetzner Cloud
	name := fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())

	// Create the key!
	key, _, err := client.SSHKey.Create(ctx, hcloud.SSHKeyCreateOpts{
		Name:      name,
		PublicKey: pubSSHFormat,
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
		if _, err := f.Write(pem.EncodeToMemory(&privBLK)); err != nil {
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
	// If no key id is set, then we never created it, so just return
	if s.keyId == 0 {
		return
	}

	client := state.Get("hcloudClient").(*hcloud.Client)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Deleting temporary ssh key...")
	_, err := client.SSHKey.Delete(context.TODO(), &hcloud.SSHKey{ID: s.keyId})
	if err != nil {
		log.Printf("Error cleaning up ssh key: %s", err)
		ui.Error(fmt.Sprintf(
			"Error cleaning up ssh key. Please delete the key manually: %s", err))
	}
}
