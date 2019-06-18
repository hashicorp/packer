package uhost

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepConfigVPC struct {
	VPCId string
}

func (s *stepConfigVPC) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*UCloudClient)
	ui := state.Get("ui").(packer.Ui)

	if len(s.VPCId) != 0 {
		ui.Say(fmt.Sprintf("Trying to use specified vpc %q...", s.VPCId))

		vpcSet, err := client.describeVPCById(s.VPCId)
		if err != nil {
			if isNotFoundError(err) {
				err = fmt.Errorf("the specified vpc %q not exist", s.VPCId)
				return halt(state, err, "")
			}
			return halt(state, err, "Error on querying vpc")
		}

		state.Put("vpc_id", vpcSet.VPCId)
		return multistep.ActionContinue

	}

	ui.Say(fmt.Sprintf("Trying to use default vpc..."))

	return multistep.ActionContinue
}

func (s *stepConfigVPC) Cleanup(state multistep.StateBag) {
}
