package profitbricks

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
)

type StepCreateSSHKey struct {
	Debug        bool
	DebugKeyPath string
}

func (s *StepCreateSSHKey) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	c := state.Get("config").(*Config)

	if (c.SSHKey_path == "") {
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
		state.Put("privateKey", string(pem.EncodeToMemory(&priv_blk)))
		state.Put("publicKey", string(ssh.MarshalAuthorizedKey(pub)))

		ui.Message(fmt.Sprintf("Saving key to: %s", s.DebugKeyPath))
		f, err := os.Create(s.DebugKeyPath)
		if err != nil {
			state.Put("error", fmt.Errorf("Error saving debug key: %s", err))
			return multistep.ActionHalt
		}

		f.Chmod(os.FileMode(int(0700)))
		err = pem.Encode(f, &priv_blk)
		f.Close()
		if err != nil {
			state.Put("error", fmt.Errorf("Error saving debug key: %s", err))
			return multistep.ActionHalt
		}
	} else {
		ui.Say(c.SSHKey_path)
		pemBytes, err := ioutil.ReadFile(c.SSHKey_path)

		if (err != nil) {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		block, _ := pem.Decode(pemBytes)

		priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)

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
		state.Put("privateKey", string(pem.EncodeToMemory(&priv_blk)))
		state.Put("publicKey", string(ssh.MarshalAuthorizedKey(pub)))
	}
	return multistep.ActionContinue
}

func (s *StepCreateSSHKey) Cleanup(state multistep.StateBag) {}
