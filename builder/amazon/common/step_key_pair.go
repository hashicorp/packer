package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepKeyPair struct {
	Debug                bool
	DebugKeyPath         string
	TemporaryKeyPairName string
	KeyPairName          string
	PrivateKeyFile       string

	keyName string
}

func (s *StepKeyPair) Run(state multistep.StateBag) multistep.StepAction {
	if s.PrivateKeyFile != "" {
		privateKeyBytes, err := ioutil.ReadFile(s.PrivateKeyFile)
		if err != nil {
			state.Put("error", fmt.Errorf(
				"Error loading configured private key file: %s", err))
			return multistep.ActionHalt
		}

		state.Put("keyPair", s.KeyPairName)
		state.Put("privateKey", string(privateKeyBytes))

		return multistep.ActionContinue
	}

	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	ui.Say(fmt.Sprintf("Creating temporary keypair: %s", s.TemporaryKeyPairName))
	keyResp, err := ec2conn.CreateKeyPair(&ec2.CreateKeyPairInput{
		KeyName: &s.TemporaryKeyPairName})
	if err != nil {
		state.Put("error", fmt.Errorf("Error creating temporary keypair: %s", err))
		return multistep.ActionHalt
	}

	// Set the keyname so we know to delete it later
	s.keyName = s.TemporaryKeyPairName

	// Set some state data for use in future steps
	state.Put("keyPair", s.keyName)
	state.Put("privateKey", *keyResp.KeyMaterial)

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
		if _, err := f.Write([]byte(*keyResp.KeyMaterial)); err != nil {
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
	// If we used an SSH private key file, do not go about deleting
	// keypairs
	if s.PrivateKeyFile != "" {
		return
	}

	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	// Remove the keypair
	ui.Say("Deleting temporary keypair...")
	_, err := ec2conn.DeleteKeyPair(&ec2.DeleteKeyPairInput{KeyName: &s.keyName})
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up keypair. Please delete the key manually: %s", s.keyName))
	}

	// Also remove the physical key if we're debugging.
	if s.Debug {
		if err := os.Remove(s.DebugKeyPath); err != nil {
			ui.Error(fmt.Sprintf(
				"Error removing debug key '%s': %s", s.DebugKeyPath, err))
		}
	}
}
