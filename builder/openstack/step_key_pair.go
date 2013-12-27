package openstack

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud"
	"log"
	"os"
	"runtime"
)

type StepKeyPair struct {
	Debug        bool
	DebugKeyPath string
	keyName      string
}

func (s *StepKeyPair) Run(state multistep.StateBag) multistep.StepAction {
	csp := state.Get("csp").(gophercloud.CloudServersProvider)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating temporary keypair for this instance...")
	keyName := fmt.Sprintf("packer %s", uuid.TimeOrderedUUID())
	log.Printf("temporary keypair name: %s", keyName)
	keyResp, err := csp.CreateKeyPair(gophercloud.NewKeyPair{Name: keyName})
	if err != nil {
		state.Put("error", fmt.Errorf("Error creating temporary keypair: %s", err))
		return multistep.ActionHalt
	}

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
		if _, err := f.Write([]byte(keyResp.PrivateKey)); err != nil {
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
	state.Put("privateKey", keyResp.PrivateKey)

	return multistep.ActionContinue
}

func (s *StepKeyPair) Cleanup(state multistep.StateBag) {
	// If no key name is set, then we never created it, so just return
	if s.keyName == "" {
		return
	}

	csp := state.Get("csp").(gophercloud.CloudServersProvider)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting temporary keypair...")
	err := csp.DeleteKeyPair(s.keyName)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up keypair. Please delete the key manually: %s", s.keyName))
	}
}
