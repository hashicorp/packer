package qemu

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"time"
)

// stepBootWait waits the configured time period.
type stepBootWait struct{}

func (s *stepBootWait) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	if int64(config.bootWait) > 0 {
		ui.Say(fmt.Sprintf("Waiting %s for boot...", config.bootWait))
		time.Sleep(config.bootWait)
	}

	return multistep.ActionContinue
}

func (s *stepBootWait) Cleanup(state multistep.StateBag) {}
