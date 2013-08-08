package chroot

import (
	"github.com/mitchellh/multistep"
)

// This step sets the various builder variables available for templates.
//
// Uses:
//   http_port int
type StepBuilderVars struct{}

func (s *StepBuilderVars) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*Config)
	device := state["device"].(string)

	config.template.BuilderVars["device"] = device

	return multistep.ActionContinue
}

func (s *StepBuilderVars) Cleanup(map[string]interface{}) {}
