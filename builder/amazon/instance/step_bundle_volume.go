package instance

import (
	"github.com/mitchellh/multistep"
	"time"
)

type StepBundleVolume struct{}

func (s *StepBundleVolume) Run(state map[string]interface{}) multistep.StepAction {
	time.Sleep(10 * time.Hour)
	return multistep.ActionContinue
}

func (s *StepBundleVolume) Cleanup(map[string]interface{}) {}
