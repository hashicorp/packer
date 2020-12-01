package common

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/antihax/optional"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/outscale/osc-sdk-go/osc"
)

type StepKeyPair struct {
	Debug        bool
	Comm         *communicator.Config
	DebugKeyPath string

	doCleanup bool
}

func (s *StepKeyPair) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
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
		ui.Say("Using SSH Agent with key pair in Source OMI")
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

	conn := state.Get("osc").(*osc.APIClient)

	ui.Say(fmt.Sprintf("Creating temporary keypair: %s", s.Comm.SSHTemporaryKeyPairName))

	resp, _, err := conn.KeypairApi.CreateKeypair(context.Background(), &osc.CreateKeypairOpts{
		CreateKeypairRequest: optional.NewInterface(osc.CreateKeypairRequest{
			KeypairName: s.Comm.SSHTemporaryKeyPairName,
		}),
	})

	if err != nil {
		state.Put("error", fmt.Errorf("Error creating temporary keypair: %s", err))
		return multistep.ActionHalt
	}

	s.doCleanup = true

	// Set some data for use in future steps
	s.Comm.SSHKeyPairName = s.Comm.SSHTemporaryKeyPairName
	s.Comm.SSHPrivateKey = []byte(resp.Keypair.PrivateKey)

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
		if _, err := f.Write([]byte(resp.Keypair.PrivateKey)); err != nil {
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

	var (
		conn = state.Get("osc").(*osc.APIClient)
		ui   = state.Get("ui").(packersdk.Ui)
	)

	// Remove the keypair
	ui.Say("Deleting temporary keypair...")
	_, _, err := conn.KeypairApi.DeleteKeypair(context.Background(), &osc.DeleteKeypairOpts{
		DeleteKeypairRequest: optional.NewInterface(osc.DeleteKeypairRequest{
			KeypairName: s.Comm.SSHTemporaryKeyPairName,
		}),
	})

	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up keypair. Please delete the key manually: %s", s.Comm.SSHTemporaryKeyPairName))
	}

	// Also remove the physical key if we're debugging.
	if s.Debug {
		if err := os.Remove(s.DebugKeyPath); err != nil {
			ui.Error(fmt.Sprintf(
				"Error removing debug key '%s': %s", s.DebugKeyPath, err))
		}
	}
}
