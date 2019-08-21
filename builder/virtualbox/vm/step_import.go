package vm

import (
	"context"

	"github.com/hashicorp/packer/helper/multistep"
)

// This step imports an OVF VM into VirtualBox.
type StepImport struct {
	Name string
}

func (s *StepImport) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	state.Put("vmName", s.Name)
	return multistep.ActionContinue
}

func (s *StepImport) Cleanup(state multistep.StateBag) {
}
