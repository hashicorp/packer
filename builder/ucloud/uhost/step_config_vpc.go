package uhost

import (
	"context"
	"fmt"

	ucloudcommon "github.com/hashicorp/packer/builder/ucloud/common"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepConfigVPC struct {
	VPCId string
}

func (s *stepConfigVPC) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ucloudcommon.UCloudClient)
	ui := state.Get("ui").(packersdk.Ui)

	if len(s.VPCId) != 0 {
		ui.Say(fmt.Sprintf("Trying to use specified vpc %q...", s.VPCId))

		vpcSet, err := client.DescribeVPCById(s.VPCId)
		if err != nil {
			if ucloudcommon.IsNotFoundError(err) {
				err = fmt.Errorf("the specified vpc %q does not exist", s.VPCId)
				return ucloudcommon.Halt(state, err, "")
			}
			return ucloudcommon.Halt(state, err, fmt.Sprintf("Error on querying specified vpc %q", s.VPCId))
		}

		state.Put("vpc_id", vpcSet.VPCId)
		return multistep.ActionContinue

	}

	ui.Say(fmt.Sprintf("Trying to use default vpc..."))

	return multistep.ActionContinue
}

func (s *stepConfigVPC) Cleanup(state multistep.StateBag) {
}
