package ecs

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepConfigAlicloudKeyPair struct {
	Debug                bool
	SSHAgentAuth         bool
	DebugKeyPath         string
	TemporaryKeyPairName string
	KeyPairName          string
	PrivateKeyFile       string
	RegionId             string

	keyName string
}

func (s *StepConfigAlicloudKeyPair) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
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
		ui.Say("Using SSH Agent with key pair in source image")
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

	client := state.Get("client").(*ecs.Client)

	ui.Say(fmt.Sprintf("Creating temporary keypair: %s", s.TemporaryKeyPairName))
	keyResp, err := client.CreateKeyPair(&ecs.CreateKeyPairArgs{
		KeyPairName: s.TemporaryKeyPairName,
		RegionId:    common.Region(s.RegionId),
	})
	if err != nil {
		state.Put("error", fmt.Errorf("Error creating temporary keypair: %s", err))
		return multistep.ActionHalt
	}

	// Set the keyname so we know to delete it later
	s.keyName = s.TemporaryKeyPairName

	// Set some state data for use in future steps
	state.Put("keyPair", s.keyName)
	state.Put("privateKey", keyResp.PrivateKeyBody)

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
		if _, err := f.Write([]byte(keyResp.PrivateKeyBody)); err != nil {
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

func (s *StepConfigAlicloudKeyPair) Cleanup(state multistep.StateBag) {
	// If no key name is set, then we never created it, so just return
	// If we used an SSH private key file, do not go about deleting
	// keypairs
	if s.PrivateKeyFile != "" || (s.KeyPairName == "" && s.keyName == "") {
		return
	}

	client := state.Get("client").(*ecs.Client)
	ui := state.Get("ui").(packer.Ui)

	// Remove the keypair
	ui.Say("Deleting temporary keypair...")
	err := client.DeleteKeyPairs(&ecs.DeleteKeyPairsArgs{
		RegionId:     common.Region(s.RegionId),
		KeyPairNames: "[\"" + s.keyName + "\"]",
	})
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
