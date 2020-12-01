package hyperone

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"runtime"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"golang.org/x/crypto/ssh"
)

type stepCreateSSHKey struct {
	Debug        bool
	DebugKeyPath string
}

func (s *stepCreateSSHKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	ui.Say("Creating a temporary ssh key for the VM...")

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		state.Put("error", fmt.Errorf("error generating ssh key: %s", err))
		return multistep.ActionHalt
	}

	privDER := x509.MarshalPKCS1PrivateKey(priv)
	privBLK := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	c.Comm.SSHPrivateKey = pem.EncodeToMemory(&privBLK)

	pub, err := ssh.NewPublicKey(&priv.PublicKey)
	if err != nil {
		state.Put("error", fmt.Errorf("error getting public key: %s", err))
		return multistep.ActionHalt
	}

	pubSSHFormat := string(ssh.MarshalAuthorizedKey(pub))

	// Remember public SSH key for future connections
	state.Put("ssh_public_key", pubSSHFormat)

	// If we're in debug mode, output the private key to the working directory.
	if s.Debug {
		ui.Message(fmt.Sprintf("Saving key for debug purposes: %s", s.DebugKeyPath))
		f, err := os.Create(s.DebugKeyPath)
		if err != nil {
			state.Put("error", fmt.Errorf("error saving debug key: %s", err))
			return multistep.ActionHalt
		}
		defer f.Close()

		// Write the key out
		if _, err := f.Write(pem.EncodeToMemory(&privBLK)); err != nil {
			state.Put("error", fmt.Errorf("error saving debug key: %s", err))
			return multistep.ActionHalt
		}

		// Chmod it so that it is SSH ready
		if runtime.GOOS != "windows" {
			if err := f.Chmod(0600); err != nil {
				state.Put("error", fmt.Errorf("error setting permissions of debug key: %s", err))
				return multistep.ActionHalt
			}
		}
	}

	return multistep.ActionContinue
}

func (s *stepCreateSSHKey) Cleanup(state multistep.StateBag) {}
