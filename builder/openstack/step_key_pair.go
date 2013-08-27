package openstack

import (
	"cgl.tideland.biz/identifier"
	"encoding/hex"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/rackspace/gophercloud"
	"log"
)

type StepKeyPair struct {
	keyName string
}

func (s *StepKeyPair) Run(state map[string]interface{}) multistep.StepAction {
	accessor := state["accessor"].(*gophercloud.Access)
	api := state["api"].(*gophercloud.ApiCriteria)
	ui := state["ui"].(packer.Ui)

	ui.Say("Creating temporary keypair for this instance...")
	keyName := fmt.Sprintf("packer %s", hex.EncodeToString(identifier.NewUUID().Raw()))
	log.Printf("temporary keypair name: %s", keyName)
	csp, err := gophercloud.ServersApi(accessor, *api)
	keyResp, err := csp.CreateKeyPair(gophercloud.NewKeyPair{Name: keyName})
	if err != nil {
		state["error"] = fmt.Errorf("Error creating temporary keypair: %s", err)
		return multistep.ActionHalt
	}

	// Set the keyname so we know to delete it later
	s.keyName = keyName

	// Set some state data for use in future steps
	state["keyPair"] = keyName
	state["privateKey"] = keyResp.PrivateKey

	return multistep.ActionContinue
}

func (s *StepKeyPair) Cleanup(state map[string]interface{}) {
	// If no key name is set, then we never created it, so just return
	if s.keyName == "" {
		return
	}

	accessor := state["accessor"].(*gophercloud.Access)
	api := state["api"].(*gophercloud.ApiCriteria)
	ui := state["ui"].(packer.Ui)

	ui.Say("Deleting temporary keypair...")
	csp, err := gophercloud.ServersApi(accessor, *api)
	err = csp.DeleteKeyPair(s.keyName)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up keypair. Please delete the key manually: %s", s.keyName))
	}
}
