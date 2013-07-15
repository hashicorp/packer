package ebs

import (
	"cgl.tideland.biz/identifier"
	"encoding/hex"
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

type stepKeyPair struct {
	keyName string
}

func (s *stepKeyPair) Run(state map[string]interface{}) multistep.StepAction {
	ec2conn := state["ec2"].(*ec2.EC2)
	ui := state["ui"].(packer.Ui)

	ui.Say("Creating temporary keypair for this instance...")
	keyName := fmt.Sprintf("packer %s", hex.EncodeToString(identifier.NewUUID().Raw()))
	log.Printf("temporary keypair name: %s", keyName)
	keyResp, err := ec2conn.CreateKeyPair(keyName)
	if err != nil {
		err := fmt.Errorf("Error creating temporary keypair: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the keyname so we know to delete it later
	s.keyName = keyName

	// Set some state data for use in future steps
	state["keyPair"] = keyName
	state["privateKey"] = keyResp.KeyMaterial

	return multistep.ActionContinue
}

func (s *stepKeyPair) Cleanup(state map[string]interface{}) {
	// If no key name is set, then we never created it, so just return
	if s.keyName == "" {
		return
	}

	ec2conn := state["ec2"].(*ec2.EC2)
	ui := state["ui"].(packer.Ui)

	ui.Say("Deleting temporary keypair...")
	_, err := ec2conn.DeleteKeyPair(s.keyName)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up keypair. Please delete the key manually: %s", s.keyName))
	}
}
