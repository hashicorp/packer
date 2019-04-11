package cvm

import (
	"context"
	"time"

	"fmt"

	"github.com/hashicorp/packer/common/retry"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/pkg/errors"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

type stepConfigSecurityGroup struct {
	SecurityGroupId   string
	SecurityGroupName string
	Description       string
	isCreate          bool
}

func (s *stepConfigSecurityGroup) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	vpcClient := state.Get("vpc_client").(*vpc.Client)
	ui := state.Get("ui").(packer.Ui)

	if len(s.SecurityGroupId) != 0 { // use existing security group
		req := vpc.NewDescribeSecurityGroupsRequest()
		req.SecurityGroupIds = []*string{&s.SecurityGroupId}
		resp, err := vpcClient.DescribeSecurityGroups(req)
		if err != nil {
			ui.Error(fmt.Sprintf("query security group failed: %s", err.Error()))
			state.Put("error", err)
			return multistep.ActionHalt
		}
		if *resp.Response.TotalCount > 0 {
			state.Put("security_group_id", s.SecurityGroupId)
			s.isCreate = false
			return multistep.ActionContinue
		}
		message := fmt.Sprintf("the specified security group(%s) does not exist", s.SecurityGroupId)
		ui.Error(message)
		state.Put("error", errors.New(message))
		return multistep.ActionHalt
	}
	// create a new security group
	req := vpc.NewCreateSecurityGroupRequest()
	req.GroupName = &s.SecurityGroupName
	req.GroupDescription = &s.Description
	resp, err := vpcClient.CreateSecurityGroup(req)
	if err != nil {
		ui.Error(fmt.Sprintf("create security group failed: %s", err.Error()))
		state.Put("error", err)
		return multistep.ActionHalt
	}
	s.SecurityGroupId = *resp.Response.SecurityGroup.SecurityGroupId
	state.Put("security_group_id", s.SecurityGroupId)
	s.isCreate = true

	// bind security group ingress police
	pReq := vpc.NewCreateSecurityGroupPoliciesRequest()
	ACCEPT := "ACCEPT"
	DEFAULT_CIDR := "0.0.0.0/0"
	pReq.SecurityGroupId = &s.SecurityGroupId
	pReq.SecurityGroupPolicySet = &vpc.SecurityGroupPolicySet{
		Ingress: []*vpc.SecurityGroupPolicy{
			{
				CidrBlock: &DEFAULT_CIDR,
				Action:    &ACCEPT,
			},
		},
	}
	_, err = vpcClient.CreateSecurityGroupPolicies(pReq)
	if err != nil {
		ui.Error(fmt.Sprintf("bind security group police failed: %s", err.Error()))
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// bind security group engress police
	pReq = vpc.NewCreateSecurityGroupPoliciesRequest()
	pReq.SecurityGroupId = &s.SecurityGroupId
	pReq.SecurityGroupPolicySet = &vpc.SecurityGroupPolicySet{
		Egress: []*vpc.SecurityGroupPolicy{
			{
				CidrBlock: &DEFAULT_CIDR,
				Action:    &ACCEPT,
			},
		},
	}
	_, err = vpcClient.CreateSecurityGroupPolicies(pReq)
	if err != nil {
		ui.Error(fmt.Sprintf("bind security group police failed: %s", err.Error()))
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepConfigSecurityGroup) Cleanup(state multistep.StateBag) {
	if !s.isCreate {
		return
	}
	ctx := context.TODO()
	vpcClient := state.Get("vpc_client").(*vpc.Client)
	ui := state.Get("ui").(packer.Ui)

	MessageClean(state, "VPC")
	req := vpc.NewDeleteSecurityGroupRequest()
	req.SecurityGroupId = &s.SecurityGroupId
	err := retry.Config{
		Tries:      60,
		RetryDelay: (&retry.Backoff{InitialBackoff: 5 * time.Second, MaxBackoff: 5 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		_, err := vpcClient.DeleteSecurityGroup(req)
		return err
	})
	if err != nil {
		ui.Error(fmt.Sprintf("delete security group(%s) failed: %s, you need to delete it by hand",
			s.SecurityGroupId, err.Error()))
		return
	}
}
