package cloudstack

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

type stepKeypair struct {
	Debug                bool
	DebugKeyPath         string
	KeyPair              string
	PrivateKeyFile       string
	SSHAgentAuth         bool
	TemporaryKeyPairName string
}

func (s *stepKeypair) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	if s.PrivateKeyFile != "" {
		privateKeyBytes, err := ioutil.ReadFile(s.PrivateKeyFile)
		if err != nil {
			state.Put("error", fmt.Errorf(
				"Error loading configured private key file: %s", err))
			return multistep.ActionHalt
		}

		state.Put("keypair", s.KeyPair)
		state.Put("privateKey", string(privateKeyBytes))

		return multistep.ActionContinue
	}

	if s.SSHAgentAuth && s.KeyPair == "" {
		ui.Say("Using SSH Agent with keypair in Source image")
		return multistep.ActionContinue
	}

	if s.SSHAgentAuth && s.KeyPair != "" {
		ui.Say(fmt.Sprintf("Using SSH Agent for existing keypair %s", s.KeyPair))
		state.Put("keypair", s.KeyPair)
		return multistep.ActionContinue
	}

	if s.TemporaryKeyPairName == "" {
		ui.Say("Not using a keypair")
		state.Put("keypair", "")
		return multistep.ActionContinue
	}

	client := state.Get("client").(*cloudstack.CloudStackClient)

	ui.Say(fmt.Sprintf("Creating temporary keypair: %s ...", s.TemporaryKeyPairName))

	p := client.SSH.NewCreateSSHKeyPairParams(s.TemporaryKeyPairName)
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

	ui.Say(fmt.Sprintf("Created temporary keypair: %s", s.TemporaryKeyPairName))

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

	// Set some state data for use in future steps
	state.Put("keypair", s.TemporaryKeyPairName)
	state.Put("privateKey", keypair.Privatekey)

	return multistep.ActionContinue
}

func (s *stepKeypair) Cleanup(state multistep.StateBag) {
	if s.TemporaryKeyPairName == "" {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	client := state.Get("client").(*cloudstack.CloudStackClient)

	ui.Say(fmt.Sprintf("Deleting temporary keypair: %s ...", s.TemporaryKeyPairName))

	_, err := client.SSH.DeleteSSHKeyPair(client.SSH.NewDeleteSSHKeyPairParams(
		s.TemporaryKeyPairName,
	))
	if err != nil {
		ui.Error(err.Error())
		ui.Error(fmt.Sprintf(
			"Error cleaning up keypair. Please delete the key manually: %s", s.TemporaryKeyPairName))
	}
}
