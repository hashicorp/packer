package proxmox

import (
	"context"

	"github.com/hashicorp/packer/helper/multistep"
)

// stepSuccess runs after the full build has succeeded.
//
// It sets the success state, which ensures cleanup does not remove the finished template
type stepSuccess struct{}

func (s *stepSuccess) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// We need to ensure stepStartVM.Cleanup doesn't delete the template (no
	// difference between VMs and templates when deleting)
	state.Put("success", true)

	return multistep.ActionContinue
}

func (s *stepSuccess) Cleanup(state multistep.StateBag) {}
