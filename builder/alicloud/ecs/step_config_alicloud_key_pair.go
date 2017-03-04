package ecs

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
)

type StepConfigAlicloudKeyPair struct {
	Debug                bool
	TemporaryKeyPairName string
	KeyPairName          string
	PrivateKeyFile       string
	PublicKeyFile        string
	SSHAgentAuth         bool

	keyName string
}

func (s *StepConfigAlicloudKeyPair) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	if s.PrivateKeyFile != "" {
		ui.Say("Using existing SSH private key")
		privateKeyBytes, err := ioutil.ReadFile(s.PrivateKeyFile)
		if err != nil {
			state.Put("error", fmt.Errorf(
				"Error loading configured private key file: %s", err))
			return multistep.ActionHalt
		}
		if s.PublicKeyFile == "" {
			s.PublicKeyFile = s.PrivateKeyFile + ".pub"
		}
		publicKeyBytes, err := ioutil.ReadFile(s.PublicKeyFile)

		state.Put("keyPair", s.KeyPairName)
		state.Put("privateKey", string(privateKeyBytes))
		state.Put("publickKey", string(publicKeyBytes))
		return multistep.ActionContinue
	}

	if s.SSHAgentAuth && s.KeyPairName == "" {
		ui.Say("Using SSH Agent with key pair in Alicloud Source Image")
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

	keyPair, err := NewKeyPair()
	if err != nil {
		ui.Say("create temporary keypair failed")
		return multistep.ActionHalt
	}
	state.Put("keyPair", s.TemporaryKeyPairName)
	if err != nil {
		state.Put("error", fmt.Errorf(
			"Error loading configured private key file: %s", err))
		return multistep.ActionHalt
	}
	state.Put("privateKey", keyPair.PrivateKey)
	state.Put("publickKey", keyPair.PublicKey)

	return multistep.ActionContinue
}

func (s *StepConfigAlicloudKeyPair) Cleanup(state multistep.StateBag) {
}
