package profitbricks

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
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

	if c.Comm.SSHPrivateKey != "" {
		pemBytes, err := ioutil.ReadFile(c.Comm.SSHPrivateKey)

		if err != nil {
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		block, _ := pem.Decode(pemBytes)

		priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)

		if err != nil {

			state.Put("error", err.Error())
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
