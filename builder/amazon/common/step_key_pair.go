package common

import (
	"cgl.tideland.biz/identifier"
	"encoding/hex"
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"os"
	"runtime"
)

type StepKeyPair struct {
	Debug        bool
	DebugKeyPath string

	keyName string
}

func (s *StepKeyPair) Run(state map[string]interface{}) multistep.StepAction {
	ec2conn := state["ec2"].(*ec2.EC2)
	ui := state["ui"].(packer.Ui)

	ui.Say("Creating temporary keypair for this instance...")
	keyName := fmt.Sprintf("packer %s", hex.EncodeToString(identifier.NewUUID().Raw()))
	log.Printf("temporary keypair name: %s", keyName)
	keyResp, err := ec2conn.CreateKeyPair(keyName)
	if err != nil {
		state["error"] = fmt.Errorf("Error creating temporary keypair: %s", err)
		return multistep.ActionHalt
	}

	// Set the keyname so we know to delete it later
	s.keyName = keyName

	// Set some state data for use in future steps
	state["keyPair"] = keyName
	state["privateKey"] = keyResp.KeyMaterial

	// If we're in debug mode, output the private key to the working
	// directory.
	if s.Debug {
		ui.Message(fmt.Sprintf("Saving key for debug purposes: %s", s.DebugKeyPath))
		f, err := os.Create(s.DebugKeyPath)
		if err != nil {
			state["error"] = fmt.Errorf("Error saving debug key: %s", err)
			return multistep.ActionHalt
		}
		defer f.Close()

		// Write the key out
		if _, err := f.Write([]byte(keyResp.KeyMaterial)); err != nil {
			state["error"] = fmt.Errorf("Error saving debug key: %s", err)
			return multistep.ActionHalt
		}

		// Chmod it so that it is SSH ready
		if runtime.GOOS != "windows" {
			if err := f.Chmod(0600); err != nil {
				state["error"] = fmt.Errorf("Error setting permissions of debug key: %s", err)
				return multistep.ActionHalt
			}
		}
	}

	return multistep.ActionContinue
}

func (s *StepKeyPair) Cleanup(state map[string]interface{}) {
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
