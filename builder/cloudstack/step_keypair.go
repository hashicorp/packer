package cloudstack

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

type stepKeypair struct {
	Debug        bool
	Comm         *communicator.Config
	DebugKeyPath string
}

func (s *stepKeypair) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	if s.Comm.SSHPrivateKeyFile != "" {
		ui.Say("Using existing SSH private key")
		privateKeyBytes, err := s.Comm.ReadSSHPrivateKeyFile()
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}

		s.Comm.SSHPrivateKey = privateKeyBytes

		return multistep.ActionContinue
	}

	if s.Comm.SSHAgentAuth && s.Comm.SSHKeyPairName == "" {
		ui.Say("Using SSH Agent with keypair in Source image")
		return multistep.ActionContinue
	}

	if s.Comm.SSHAgentAuth && s.Comm.SSHKeyPairName != "" {
		ui.Say(fmt.Sprintf("Using SSH Agent for existing keypair %s", s.Comm.SSHKeyPairName))
		return multistep.ActionContinue
	}

	if s.Comm.SSHTemporaryKeyPairName == "" {
		ui.Say("Not using a keypair")
		s.Comm.SSHKeyPairName = ""
		return multistep.ActionContinue
	}

	client := state.Get("client").(*cloudstack.CloudStackClient)

	ui.Say(fmt.Sprintf("Creating temporary keypair: %s ...", s.Comm.SSHTemporaryKeyPairName))

	p := client.SSH.NewCreateSSHKeyPairParams(s.Comm.SSHTemporaryKeyPairName)

	cfg := state.Get("config").(*Config)
	if cfg.Project != "" {
		p.SetProjectid(cfg.Project)
	}

	keypair, err := client.SSH.CreateSSHKeyPair(p)
	if err != nil {
		err := fmt.Errorf("Error creating temporary keypair: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if keypair.Privatekey == "" {
		err := fmt.Errorf("The temporary keypair returned was blank")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Created temporary keypair: %s", s.Comm.SSHTemporaryKeyPairName))

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
		if _, err := f.Write([]byte(keypair.Privatekey)); err != nil {
			err := fmt.Errorf("Error saving debug key: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Chmod it so that it is SSH ready
		if runtime.GOOS != "windows" {
			if err := f.Chmod(0600); err != nil {
				err := fmt.Errorf("Error setting permissions of debug key: %s", err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	// Set some data for use in future steps
	s.Comm.SSHKeyPairName = s.Comm.SSHTemporaryKeyPairName
	s.Comm.SSHPrivateKey = []byte(keypair.Privatekey)

	return multistep.ActionContinue
}

func (s *stepKeypair) Cleanup(state multistep.StateBag) {
	if s.Comm.SSHTemporaryKeyPairName == "" {
		return
	}

	ui := state.Get("ui").(packersdk.Ui)
	client := state.Get("client").(*cloudstack.CloudStackClient)
	cfg := state.Get("config").(*Config)

	p := client.SSH.NewDeleteSSHKeyPairParams(s.Comm.SSHTemporaryKeyPairName)
	if cfg.Project != "" {
		p.SetProjectid(cfg.Project)
	}

	ui.Say(fmt.Sprintf("Deleting temporary keypair: %s ...", s.Comm.SSHTemporaryKeyPairName))

	_, err := client.SSH.DeleteSSHKeyPair(p)
	if err != nil {
		ui.Error(err.Error())
		ui.Error(fmt.Sprintf(
			"Error cleaning up keypair. Please delete the key manually: %s", s.Comm.SSHTemporaryKeyPairName))
	}
}
