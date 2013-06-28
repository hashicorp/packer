package virtualbox

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"os"
)

type stepPrepareOutputDir struct{}

func (stepPrepareOutputDir) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)

	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		state["error"] = err
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (stepPrepareOutputDir) Cleanup(state map[string]interface{}) {
	_, cancelled := state[multistep.StateCancelled]
	_, halted := state[multistep.StateHalted]

	if cancelled || halted {
		config := state["config"].(*config)
		ui := state["ui"].(packer.Ui)

		ui.Say("Deleting output directory...")
		os.RemoveAll(config.OutputDir)
	}
}
