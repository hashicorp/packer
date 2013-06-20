package digitalocean

import (
	"cgl.tideland.biz/identifier"
	"code.google.com/p/go.crypto/ssh"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

type stepCreateSSHKey struct {
	keyId uint
}

func (s *stepCreateSSHKey) Run(state map[string]interface{}) multistep.StepAction {
	client := state["client"].(*DigitalOceanClient)
	ui := state["ui"].(packer.Ui)

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
	state["privateKey"] = string(pem.EncodeToMemory(&priv_blk))

	// Marshal the public key into SSH compatible format
	pub := priv.PublicKey
	pub_sshformat := string(ssh.MarshalAuthorizedKey(&pub))

	// The name of the public key on DO
	name := fmt.Sprintf("packer-%s", hex.EncodeToString(identifier.NewUUID().Raw()))

	// Create the key!
	keyId, err := client.CreateKey(name, pub_sshformat)
	if err != nil {
		err := fmt.Errorf("Error creating temporary SSH key: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// We use this to check cleanup
	s.keyId = keyId

	log.Printf("temporary ssh key name: %s", name)

	// Remember some state for the future
	state["ssh_key_id"] = keyId

	return multistep.ActionContinue
}

func (s *stepCreateSSHKey) Cleanup(state map[string]interface{}) {
	// If no key name is set, then we never created it, so just return
	if s.keyId == 0 {
		return
	}

	client := state["client"].(*DigitalOceanClient)
	ui := state["ui"].(packer.Ui)
	c := state["config"].(config)

	ui.Say("Deleting temporary ssh key...")
	err := client.DestroyKey(s.keyId)

	curlstr := fmt.Sprintf("curl '%v/ssh_keys/%v/destroy?client_id=%v&api_key=%v'",
		DIGITALOCEAN_API_URL, s.keyId, c.ClientID, c.APIKey)

	if err != nil {
		log.Printf("Error cleaning up ssh key: %v", err.Error())
		ui.Error(fmt.Sprintf(
			"Error cleaning up ssh key. Please delete the key manually: %v", curlstr))
	}
}
