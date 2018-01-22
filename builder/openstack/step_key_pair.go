package openstack

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/keypairs"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"golang.org/x/crypto/ssh"
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
		ui.Say("Using SSH Agent with key pair in Source image")
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

	config := state.Get("config").(Config)

	// We need the v2 compute client
	computeClient, err := config.computeV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing compute client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Creating temporary keypair: %s ...", s.TemporaryKeyPairName))
	keypair, err := keypairs.Create(computeClient, keypairs.CreateOpts{
		Name: s.TemporaryKeyPairName,
	}).Extract()
	if err != nil {
		state.Put("error", fmt.Errorf("Error creating temporary keypair: %s", err))
		return multistep.ActionHalt
	}

	if keypair.PrivateKey == "" {
		state.Put("error", fmt.Errorf("The temporary keypair returned was blank"))
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Created temporary keypair: %s", s.TemporaryKeyPairName))

	keypair.PrivateKey = berToDer(keypair.PrivateKey, ui)

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
	state.Put("keyPair", s.TemporaryKeyPairName)
	state.Put("privateKey", keypair.PrivateKey)

	return multistep.ActionContinue
}

// Work around for https://github.com/hashicorp/packer/issues/2526
func berToDer(ber string, ui packer.Ui) string {
	// Check if x/crypto/ssh can parse the key
	_, err := ssh.ParsePrivateKey([]byte(ber))
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

	berKey, err := ioutil.TempFile("", "packer-ber-privatekey-")
	defer os.Remove(berKey.Name())
	if err != nil {
		return ber
	}
	ioutil.WriteFile(berKey.Name(), []byte(ber), os.ModeAppend)
	derKey, err := ioutil.TempFile("", "packer-der-privatekey-")
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
	return string(der)
}

func (s *StepKeyPair) Cleanup(state multistep.StateBag) {
	if !s.doCleanup {
		return
	}

	config := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)

	// We need the v2 compute client
	computeClient, err := config.computeV2Client()
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up keypair. Please delete the key manually: %s", s.TemporaryKeyPairName))
		return
	}

	ui.Say(fmt.Sprintf("Deleting temporary keypair: %s ...", s.TemporaryKeyPairName))
	err = keypairs.Delete(computeClient, s.TemporaryKeyPairName).ExtractErr()
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up keypair. Please delete the key manually: %s", s.TemporaryKeyPairName))
	}
}
