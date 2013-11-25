package docker

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

type StepPull struct{}

func (s *StepPull) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	if !config.Pull {
		log.Println("Pull disabled, won't docker pull")
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Pulling Docker image: %s", config.Image))
	if err := driver.Pull(config.Image); err != nil {
		err := fmt.Errorf("Error pulling Docker image: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepPull) Cleanup(state multistep.StateBag) {
}
