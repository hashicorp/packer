package uhost

import (
	"context"
	"fmt"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
)

type stepConfigSecurityGroup struct {
	SecurityGroupId string
}

func (s *stepConfigSecurityGroup) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*UCloudClient)
	conn := client.unetconn
	ui := state.Get("ui").(packer.Ui)

	if len(s.SecurityGroupId) != 0 {
		ui.Say(fmt.Sprintf("Trying to use specified security group %q...", s.SecurityGroupId))
		securityGroupSet, err := client.describeFirewallById(s.SecurityGroupId)
		if err != nil {
			if isNotFoundError(err) {
				err = fmt.Errorf("the specified security group %q not exist", s.SecurityGroupId)
				return halt(state, err, "")
			}
			return halt(state, err, "Error on querying security group")
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
			return halt(state, err, "Error on querying security group")
		}

		if resp == nil || len(resp.DataSet) < 1 {
			break
		}

		for _, item := range resp.DataSet {
			if item.Type == securityGroupNonWeb {
				securityGroupId = item.FWId
				break
			}
		}

		if len(resp.DataSet) < limit {
			break
		}

		offset = offset + limit
	}

	if securityGroupId != "" {
		state.Put("security_group_id", securityGroupId)
		return multistep.ActionContinue
	}
	return multistep.ActionContinue
}

func (s *stepConfigSecurityGroup) Cleanup(state multistep.StateBag) {
}
