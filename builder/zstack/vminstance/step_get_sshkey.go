package vminstance

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/packer/helper/multistep"
)

type StepGetSSHKey struct {
	Publicfile string
}

func (s *StepGetSSHKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	_, config, ui := GetCommonFromState(state)
	ui.Say("start get sshkey for zstack...")

	if s.Publicfile != "" {
		publicKeyBytes, err := ioutil.ReadFile(s.Publicfile)
		if err != nil {
			state.Put("error", fmt.Errorf(
				"Error loading configured public key file: %s", err))
			return multistep.ActionHalt
		}
		config.Comm.SSHPublicKey = publicKeyBytes
		state.Put("config", config)
	}

	return multistep.ActionContinue
}

func (s *StepGetSSHKey) Cleanup(state multistep.StateBag) {
	_, _, ui := GetCommonFromState(state)
	ui.Say("cleanup get sshkey executing...")
}
