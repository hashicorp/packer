package digitalocean

import (
	"cgl.tideland.biz/identifier"
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
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the pem formatted private key on the state for later
	priv_der := x509.MarshalPKCS1PrivateKey(priv)
	priv_blk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   priv_der,
	}

	// Create the public key for uploading to DO
	pub := priv.PublicKey
	pub_der, err := x509.MarshalPKIXPublicKey(&pub)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	pub_blk := pem.Block{
		Type:    "PUBLIC KEY",
		Headers: nil,
		Bytes:   pub_der,
	}
	pub_pem := string(pem.EncodeToMemory(&pub_blk))

	name := fmt.Sprintf("packer %s", hex.EncodeToString(identifier.NewUUID().Raw()))

	keyId, err := client.CreateKey(name, pub_pem)

	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// We use this to check cleanup
	s.keyId = keyId

	log.Printf("temporary ssh key name: %s", name)

	// Remember some state for the future
	state["keyId"] = keyId
	state["privateKey"] = string(pem.EncodeToMemory(&priv_blk))

	return multistep.ActionContinue
}

func (s *stepCreateSSHKey) Cleanup(state map[string]interface{}) {
	// If no key name is set, then we never created it, so just return
	if s.keyId == 0 {
		return
	}

	client := state["client"].(*DigitalOceanClient)
	ui := state["ui"].(packer.Ui)

	ui.Say("Deleting temporary ssh key...")
	err := client.DestroyKey(s.keyId)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up ssh key. Please delete the key manually: %s", s.keyId))
	}
}
