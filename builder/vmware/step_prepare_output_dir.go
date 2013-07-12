package vmware

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"os"
	"time"
)

type stepPrepareOutputDir struct{}

func (stepPrepareOutputDir) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	ui := state["ui"].(packer.Ui)

	if _, err := os.Stat(config.OutputDir); err == nil && config.PackerForce {
		ui.Say("Deleting previous output directory...")
		os.RemoveAll(config.OutputDir)
	}

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
		for i := 0; i < 5; i++ {
			err := os.RemoveAll(config.OutputDir)
			if err == nil {
				break
			}

			log.Printf("Error removing output dir: %s", err)
			time.Sleep(2 * time.Second)
		}
	}
}
