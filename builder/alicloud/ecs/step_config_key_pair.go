package ecs

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepConfigAlicloudKeyPair struct {
	Debug        bool
	Comm         *communicator.Config
	DebugKeyPath string
	RegionId     string

	keyName string
}

func (s *stepConfigAlicloudKeyPair) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
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
		ui.Say("Using SSH Agent with key pair in source image")
		return multistep.ActionContinue
	}

	if s.Comm.SSHAgentAuth && s.Comm.SSHKeyPairName != "" {
		ui.Say(fmt.Sprintf("Using SSH Agent for existing key pair %s", s.Comm.SSHKeyPairName))
		return multistep.ActionContinue
	}

	if s.Comm.SSHTemporaryKeyPairName == "" {
		ui.Say("Not using temporary keypair")
		s.Comm.SSHKeyPairName = ""
		return multistep.ActionContinue
	}

	client := state.Get("client").(*ClientWrapper)
	ui.Say(fmt.Sprintf("Creating temporary keypair: %s", s.Comm.SSHTemporaryKeyPairName))

	createKeyPairRequest := ecs.CreateCreateKeyPairRequest()
	createKeyPairRequest.RegionId = s.RegionId
	createKeyPairRequest.KeyPairName = s.Comm.SSHTemporaryKeyPairName
	keyResp, err := client.CreateKeyPair(createKeyPairRequest)
	if err != nil {
		return halt(state, err, "Error creating temporary keypair")
	}

	// Set the keyname so we know to delete it later
	s.keyName = s.Comm.SSHTemporaryKeyPairName

	// Set some state data for use in future steps
	s.Comm.SSHKeyPairName = s.keyName
	s.Comm.SSHPrivateKey = []byte(keyResp.PrivateKeyBody)

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

func (s *stepConfigAlicloudKeyPair) Cleanup(state multistep.StateBag) {
	// If no key name is set, then we never created it, so just return
	// If we used an SSH private key file, do not go about deleting
	// keypairs
	if s.Comm.SSHPrivateKeyFile != "" || (s.Comm.SSHKeyPairName == "" && s.keyName == "") {
		return
	}

	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packersdk.Ui)

	// Remove the keypair
	ui.Say("Deleting temporary keypair...")

	deleteKeyPairsRequest := ecs.CreateDeleteKeyPairsRequest()
	deleteKeyPairsRequest.RegionId = s.RegionId
	deleteKeyPairsRequest.KeyPairNames = fmt.Sprintf("[\"%s\"]", s.keyName)
	_, err := client.DeleteKeyPairs(deleteKeyPairsRequest)
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
