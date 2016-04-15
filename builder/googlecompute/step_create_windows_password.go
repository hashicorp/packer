package googlecompute

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// StepCreateWindowsPassword represents a Packer build step that generates SSH key pairs.
type StepCreateWindowsPassword struct {
	Debug        bool
	DebugKeyPath string
}

// Run executes the Packer build step that generates SSH key pairs.
func (s *StepCreateWindowsPassword) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating windows user for instance...")
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		err := fmt.Errorf("Error creating temporary key: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, priv.E); err != nil {
		err := fmt.Errorf("Error creating temporary key: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	data := WindowsPasswordConfig{
		UserName: "pieter_lazzaro",
		Modulus:  base64.StdEncoding.EncodeToString(priv.N.Bytes()),
		Exponent: base64.StdEncoding.EncodeToString(buf.Bytes()),
		Email:    "pieter.lazzaro@pureharvest.com.au",
		ExpireOn: time.Now().Add(time.Minute * 5),
	}
	state.Put("windows-keys", data)

	return multistep.ActionContinue
}

// Nothing to clean up. SSH keys are associated with a single GCE instance.
func (s *StepCreateWindowsPassword) Cleanup(state multistep.StateBag) {}
