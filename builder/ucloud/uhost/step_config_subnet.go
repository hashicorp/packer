package uhost

import (
	"context"
	"fmt"

	ucloudcommon "github.com/hashicorp/packer/builder/ucloud/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepConfigSubnet struct {
	SubnetId string
}

func (s *stepConfigSubnet) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ucloudcommon.UCloudClient)
	ui := state.Get("ui").(packer.Ui)

	if len(s.SubnetId) != 0 {
		ui.Say(fmt.Sprintf("Trying to use specified subnet %q...", s.SubnetId))
		subnetSet, err := client.DescribeSubnetById(s.SubnetId)
		if err != nil {
			if ucloudcommon.IsNotFoundError(err) {
				err = fmt.Errorf("the specified subnet %q does not exist", s.SubnetId)
				return ucloudcommon.Halt(state, err, "")
			}
			return ucloudcommon.Halt(state, err, fmt.Sprintf("Error on querying specified subnet %q", s.SubnetId))
		}

		state.Put("subnet_id", subnetSet.SubnetId)
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Trying to use default subnet..."))

	return multistep.ActionContinue
}

func (s *stepConfigSubnet) Cleanup(state multistep.StateBag) {
}
