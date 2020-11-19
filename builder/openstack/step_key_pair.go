package openstack

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/tmp"
	"golang.org/x/crypto/ssh"
)

type StepKeyPair struct {
	Debug        bool
	Comm         *communicator.Config
	DebugKeyPath string

	doCleanup bool
}

func (s *StepKeyPair) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
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
		ui.Say("Using SSH Agent with key pair in Source image")
		return multistep.ActionContinue
	}

	if s.Comm.SSHAgentAuth && s.Comm.SSHKeyPairName != "" {
		ui.Say(fmt.Sprintf("Using SSH Agent for existing key pair %s", s.Comm.SSHKeyPairName))
		s.Comm.SSHKeyPairName = ""
		return multistep.ActionContinue
	}

	if s.Comm.SSHTemporaryKeyPairName == "" {
		ui.Say("Not using temporary keypair")
		s.Comm.SSHKeyPairName = ""
		return multistep.ActionContinue
	}

	config := state.Get("config").(*Config)

	// We need the v2 compute client
	computeClient, err := config.computeV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing compute client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Creating temporary keypair: %s ...", s.Comm.SSHTemporaryKeyPairName))
	keypair, err := keypairs.Create(computeClient, keypairs.CreateOpts{
		Name: s.Comm.SSHTemporaryKeyPairName,
	}).Extract()
	if err != nil {
		state.Put("error", fmt.Errorf("Error creating temporary keypair: %s", err))
		return multistep.ActionHalt
	}

	if len(keypair.PrivateKey) == 0 {
		state.Put("error", fmt.Errorf("The temporary keypair returned was blank"))
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Created temporary keypair: %s", s.Comm.SSHTemporaryKeyPairName))

	keypair.PrivateKey = string(berToDer([]byte(keypair.PrivateKey), ui))

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
		if _, err := f.Write([]byte(keypair.PrivateKey)); err != nil {
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

	// we created a temporary key, so remember to clean it up
	s.doCleanup = true

	// Set some state data for use in future steps
	s.Comm.SSHKeyPairName = s.Comm.SSHTemporaryKeyPairName
	s.Comm.SSHPrivateKey = []byte(keypair.PrivateKey)

	return multistep.ActionContinue
}

// Work around for https://github.com/hashicorp/packer/issues/2526
func berToDer(ber []byte, ui packersdk.Ui) []byte {
	// Check if x/crypto/ssh can parse the key
	_, err := ssh.ParsePrivateKey(ber)
	if err == nil {
		return ber
	}
	// Can't parse the key, maybe it's BER encoded. Try to convert it with OpenSSL.
	log.Println("Couldn't parse SSH key, trying work around for [GH-2526].")

	openSslPath, err := exec.LookPath("openssl")
	if err != nil {
		log.Println("Couldn't find OpenSSL, aborting work around.")
		return ber
	}

	berKey, err := tmp.File("packer-ber-privatekey-")
	defer os.Remove(berKey.Name())
	if err != nil {
		return ber
	}
	ioutil.WriteFile(berKey.Name(), ber, os.ModeAppend)
	derKey, err := tmp.File("packer-der-privatekey-")
	defer os.Remove(derKey.Name())
	if err != nil {
		return ber
	}

	args := []string{"rsa", "-in", berKey.Name(), "-out", derKey.Name()}
	log.Printf("Executing: %s %v", openSslPath, args)
	if err := exec.Command(openSslPath, args...).Run(); err != nil {
		log.Printf("OpenSSL failed with error: %s", err)
		return ber
	}

	der, err := ioutil.ReadFile(derKey.Name())
	if err != nil {
		return ber
	}
	ui.Say("Successfully converted BER encoded SSH key to DER encoding.")
	return der
}

func (s *StepKeyPair) Cleanup(state multistep.StateBag) {
	if !s.doCleanup {
		return
	}

	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	// We need the v2 compute client
	computeClient, err := config.computeV2Client()
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up keypair. Please delete the key manually: %s", s.Comm.SSHTemporaryKeyPairName))
		return
	}

	ui.Say(fmt.Sprintf("Deleting temporary keypair: %s ...", s.Comm.SSHTemporaryKeyPairName))
	err = keypairs.Delete(computeClient, s.Comm.SSHTemporaryKeyPairName).ExtractErr()
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up keypair. Please delete the key manually: %s", s.Comm.SSHTemporaryKeyPairName))
	}
}
