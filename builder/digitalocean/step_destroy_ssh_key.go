package digitalocean

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepDestroySSHKey struct{}

func (s *stepDestroySSHKey) Run(state map[string]interface{}) multistep.StepAction {
	client := state["client"].(*DigitalOceanClient)
	ui := state["ui"].(packer.Ui)
	sshKeyId := state["ssh_key_id"].(uint)

	ui.Say("Destroying temporary ssh key...")

	err := client.DestroyKey(sshKeyId)

	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepDestroySSHKey) Cleanup(state map[string]interface{}) {
	// no cleanup
}
