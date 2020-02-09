package uhost

import (
	"context"
	"fmt"

	ucloudcommon "github.com/hashicorp/packer/builder/ucloud/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

type stepConfigSecurityGroup struct {
	SecurityGroupId string
}

func (s *stepConfigSecurityGroup) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ucloudcommon.UCloudClient)
	conn := client.UNetConn
	ui := state.Get("ui").(packer.Ui)

	if len(s.SecurityGroupId) != 0 {
		ui.Say(fmt.Sprintf("Trying to use specified security group %q...", s.SecurityGroupId))
		securityGroupSet, err := client.DescribeFirewallById(s.SecurityGroupId)
		if err != nil {
			if ucloudcommon.IsNotFoundError(err) {
				err = fmt.Errorf("the specified security group %q does not exist", s.SecurityGroupId)
				return ucloudcommon.Halt(state, err, "")
			}
			return ucloudcommon.Halt(state, err, fmt.Sprintf("Error on querying specified security group %q", s.SecurityGroupId))
		}

		state.Put("security_group_id", securityGroupSet.FWId)
		return multistep.ActionContinue
	}

	ui.Say("Trying to use default security group...")
	var securityGroupId string
	var limit = 100
	var offset int

	for {
		req := conn.NewDescribeFirewallRequest()
		req.Limit = ucloud.Int(limit)
		req.Offset = ucloud.Int(offset)

		resp, err := conn.DescribeFirewall(req)
		if err != nil {
			return ucloudcommon.Halt(state, err, "Error on querying default security group")
		}

		if resp == nil || len(resp.DataSet) < 1 {
			break
		}

		for _, item := range resp.DataSet {
			if item.Type == ucloudcommon.SecurityGroupNonWeb {
				securityGroupId = item.FWId
				break
			}
		}

		if len(resp.DataSet) < limit {
			break
		}

		offset = offset + limit
	}

	if securityGroupId == "" {
		return ucloudcommon.Halt(state, fmt.Errorf("the default security group does not exist"), "")
	}

	state.Put("security_group_id", securityGroupId)
	return multistep.ActionContinue
}

func (s *stepConfigSecurityGroup) Cleanup(state multistep.StateBag) {
}
