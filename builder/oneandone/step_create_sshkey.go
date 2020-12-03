package oneandone

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"golang.org/x/crypto/ssh"
)

type StepCreateSSHKey struct {
	Debug        bool
	DebugKeyPath string
}

func (s *StepCreateSSHKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)

	if c.Comm.SSHPrivateKeyFile != "" {
		pemBytes, err := ioutil.ReadFile(c.Comm.SSHPrivateKeyFile)

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
		c.Comm.SSHPrivateKey = pem.EncodeToMemory(&priv_blk)
		c.Comm.SSHPublicKey = ssh.MarshalAuthorizedKey(pub)
	}
	return multistep.ActionContinue
}

func (s *StepCreateSSHKey) Cleanup(state multistep.StateBag) {}
