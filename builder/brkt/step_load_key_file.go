package brkt

import (
	"fmt"
	"io/ioutil"

	"github.com/mitchellh/multistep"
)

//
// Loading key file
//
type stepLoadKeyFile struct {
	PrivateKeyFile string
}

func (s *stepLoadKeyFile) Run(state multistep.StateBag) multistep.StepAction {
	if s.PrivateKeyFile == "" {
		state.Put("privateKeyBastion", "")
		return multistep.ActionContinue
	}

	privateKeyBytes, err := ioutil.ReadFile(s.PrivateKeyFile)
	if err != nil {
		state.Put("error", fmt.Errorf(
			"error loading configured private key file: %s", err))
		return multistep.ActionHalt
	}

	state.Put("privateKeyBastion", string(privateKeyBytes))

	return multistep.ActionContinue
}

func (s *stepLoadKeyFile) Cleanup(state multistep.StateBag) {}
