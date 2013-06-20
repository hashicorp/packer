package virtualbox

import (
	"github.com/mitchellh/multistep"
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

func (stepPrepareOutputDir) Cleanup(map[string]interface{}) {}
