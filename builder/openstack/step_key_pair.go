package openstack

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud/openstack/compute/v2/extensions/keypairs"
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

	config := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)

	// We need the v2 compute client
	computeClient, err := config.computeV2Client()
	if err != nil {
		err = fmt.Errorf("Error initializing compute client: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	keyName := fmt.Sprintf("packer %s", uuid.TimeOrderedUUID())
	ui.Say(fmt.Sprintf("Creating temporary keypair: %s ...", keyName))
	keypair, err := keypairs.Create(computeClient, keypairs.CreateOpts{
		Name: keyName,
	}).Extract()
	if err != nil {
		state.Put("error", fmt.Errorf("Error creating temporary keypair: %s", err))
		return multistep.ActionHalt
	}

	if keypair.PrivateKey == "" {
		state.Put("error", fmt.Errorf("The temporary keypair returned was blank"))
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Created temporary keypair: %s", keyName))

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

	// Set the keyname so we know to delete it later
	s.keyName = keyName

	// Set some state data for use in future steps
	state.Put("keyPair", keyName)
	state.Put("privateKey", keypair.PrivateKey)

	return multistep.ActionContinue
}

func (s *StepKeyPair) Cleanup(state multistep.StateBag) {
	// If we used an SSH private key file, do not go about deleting
	// keypairs
	if s.PrivateKeyFile != "" {
		return
	}
	// If no key name is set, then we never created it, so just return
	if s.keyName == "" {
		return
	}

	config := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)

	// We need the v2 compute client
	computeClient, err := config.computeV2Client()
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up keypair. Please delete the key manually: %s", s.keyName))
		return
	}

	ui.Say(fmt.Sprintf("Deleting temporary keypair: %s ...", s.keyName))
	err = keypairs.Delete(computeClient, s.keyName).ExtractErr()
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up keypair. Please delete the key manually: %s", s.keyName))
	}
}
