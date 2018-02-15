package common

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepKeyPair struct {
	Debug                bool
	SSHAgentAuth         bool
	DebugKeyPath         string
	TemporaryKeyPairName string
	KeyPairName          string
	PrivateKeyFile       string

	doCleanup bool
}

func (s *StepKeyPair) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	if s.PrivateKeyFile != "" {
		ui.Say("Using existing SSH private key")
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

	if s.SSHAgentAuth && s.KeyPairName == "" {
		ui.Say("Using SSH Agent with key pair in Source AMI")
		return multistep.ActionContinue
	}

	if s.SSHAgentAuth && s.KeyPairName != "" {
		ui.Say(fmt.Sprintf("Using SSH Agent for existing key pair %s", s.KeyPairName))
		state.Put("keyPair", s.KeyPairName)
		return multistep.ActionContinue
	}

	if s.TemporaryKeyPairName == "" {
		ui.Say("Not using temporary keypair")
		state.Put("keyPair", "")
		return multistep.ActionContinue
	}

	ec2conn := state.Get("ec2").(*ec2.EC2)

	ui.Say(fmt.Sprintf("Creating temporary keypair: %s", s.TemporaryKeyPairName))
	keyResp, err := ec2conn.CreateKeyPair(&ec2.CreateKeyPairInput{
		KeyName: &s.TemporaryKeyPairName})
	if err != nil {
		state.Put("error", fmt.Errorf("Error creating temporary keypair: %s", err))
		return multistep.ActionHalt
	}

	s.doCleanup = true

	// Set some state data for use in future steps
	state.Put("keyPair", s.TemporaryKeyPairName)
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
	if !s.doCleanup {
		return
	}

	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	// Remove the keypair
	ui.Say("Deleting temporary keypair...")
	_, err := ec2conn.DeleteKeyPair(&ec2.DeleteKeyPairInput{KeyName: &s.TemporaryKeyPairName})
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up keypair. Please delete the key manually: %s", s.TemporaryKeyPairName))
	}

	// Also remove the physical key if we're debugging.
	if s.Debug {
		if err := os.Remove(s.DebugKeyPath); err != nil {
			ui.Error(fmt.Sprintf(
				"Error removing debug key '%s': %s", s.DebugKeyPath, err))
		}
	}
}
