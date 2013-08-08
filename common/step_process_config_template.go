package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

// StepProcessConfigTemplate processes the given configuration template
// and shows any errors if something goes wrong.
//
// Uses:
//   ui           packer.Ui
//
// Produces:
//   <nothing>
type StepProcessConfigTemplate struct {
	ConfigTemplate *ConfigTemplate
}

func (s *StepProcessConfigTemplate) Run(state map[string]interface{}) multistep.StepAction {
	ui := state["ui"].(packer.Ui)

	log.Println("Processing configuration template...")
	if err := s.ConfigTemplate.Process(); err != nil {
		state["error"] = fmt.Errorf(
			"Error processing configuration templates:\n", err)
		ui.Error(state["error"].(error).Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (*StepProcessConfigTemplate) Cleanup(map[string]interface{}) {}
