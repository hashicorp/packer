package openstack

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud"
	"log"
)

type StepKeyPair struct {
	keyName string
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
