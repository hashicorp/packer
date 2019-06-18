package uhost

import (
	"context"
	"fmt"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepConfigSubnet struct {
	SubnetId string
}

func (s *stepConfigSubnet) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*UCloudClient)
	ui := state.Get("ui").(packer.Ui)

	if len(s.SubnetId) != 0 {
		ui.Say(fmt.Sprintf("Trying to use specified subnet %q...", s.SubnetId))
		subnetSet, err := client.describeSubnetById(s.SubnetId)
		if err != nil {
			if isNotFoundError(err) {
				err = fmt.Errorf("the specified subnet %q not exist", s.SubnetId)
				return halt(state, err, "")
			}
			return halt(state, err, "Error on querying subnet")
		}

		state.Put("subnet_id", subnetSet.SubnetId)
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Trying to use default subnet..."))

	return multistep.ActionContinue
}

func (s *stepConfigSubnet) Cleanup(state multistep.StateBag) {
}
