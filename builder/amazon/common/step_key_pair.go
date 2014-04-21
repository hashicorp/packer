package common

import (
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
	"runtime"
)

type StepKeyPair struct {
	Debug          bool
	DebugKeyPath   string
	KeyPairName    string
	PrivateKeyFile string

	keyName string
}

func (s *StepKeyPair) Run(state multistep.StateBag) multistep.StepAction {
	if s.PrivateKeyFile != "" {
		s.keyName = ""

		privateKeyBytes, err := ioutil.ReadFile(s.PrivateKeyFile)
		if err != nil {
			state.Put("error", fmt.Errorf("Error loading configured private key file: %s", err))
			return multistep.ActionHalt
		}

		state.Put("keyPair", "")
		state.Put("privateKey", string(privateKeyBytes))

		return multistep.ActionContinue
	}

	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	ui.Say(fmt.Sprintf("Creating temporary keypair: %s", s.KeyPairName))
	keyResp, err := ec2conn.CreateKeyPair(s.KeyPairName)
	if err != nil {
		state.Put("error", fmt.Errorf("Error creating temporary keypair: %s", err))
		return multistep.ActionHalt
	}

	// Set the keyname so we know to delete it later
	s.keyName = s.KeyPairName

	// Set some state data for use in future steps
	state.Put("keyPair", s.keyName)
	state.Put("privateKey", keyResp.KeyMaterial)

	// If we're in debug mode, output the private key to the working
	// directory.
	if s.Debug {
		ui.Message(fmt.Sprintf("Saving key for debug purposes: %s", s.DebugKeyPath))
		f, err := os.Create(s.DebugKeyPath)
		if err != nil {
			state.Put("error", fmt.Errorf("Error saving debug key: %s", err))
			return multistep.ActionHalt
		}
		defer f.Close()

		// Write the key out
		if _, err := f.Write([]byte(keyResp.KeyMaterial)); err != nil {
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

func (s *StepKeyPair) Cleanup(state multistep.StateBag) {
	// If no key name is set, then we never created it, so just return
	if s.keyName == "" {
		return
	}

	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting temporary keypair...")
	_, err := ec2conn.DeleteKeyPair(s.keyName)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up keypair. Please delete the key manually: %s", s.keyName))
	}
}
