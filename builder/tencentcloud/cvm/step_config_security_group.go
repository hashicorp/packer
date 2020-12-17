package cvm

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
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

	if len(s.SecurityGroupId) != 0 {
		Say(state, s.SecurityGroupId, "Trying to use existing securitygroup")
		req := vpc.NewDescribeSecurityGroupsRequest()
		req.SecurityGroupIds = []*string{&s.SecurityGroupId}
		var resp *vpc.DescribeSecurityGroupsResponse
		err := Retry(ctx, func(ctx context.Context) error {
			var e error
			resp, e = vpcClient.DescribeSecurityGroups(req)
			return e
		})
		if err != nil {
			return Halt(state, err, "Failed to get securitygroup info")
		}
		if *resp.Response.TotalCount > 0 {
			s.isCreate = false
			state.Put("security_group_id", s.SecurityGroupId)
			Message(state, *resp.Response.SecurityGroupSet[0].SecurityGroupName, "Securitygroup found")
			return multistep.ActionContinue
		}
		return Halt(state, fmt.Errorf("The specified securitygroup(%s) does not exists", s.SecurityGroupId), "")
	}

	Say(state, "Trying to create a new securitygroup", "")

	req := vpc.NewCreateSecurityGroupRequest()
	req.GroupName = &s.SecurityGroupName
	req.GroupDescription = &s.Description
	var resp *vpc.CreateSecurityGroupResponse
	err := Retry(ctx, func(ctx context.Context) error {
		var e error
		resp, e = vpcClient.CreateSecurityGroup(req)
		return e
	})
	if err != nil {
		return Halt(state, err, "Failed to create securitygroup")
	}

	s.isCreate = true
	s.SecurityGroupId = *resp.Response.SecurityGroup.SecurityGroupId
	state.Put("security_group_id", s.SecurityGroupId)
	Message(state, s.SecurityGroupId, "Securitygroup created")

	// bind securitygroup ingress police
	Say(state, "Trying to create securitygroup polices", "")
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
	err = Retry(ctx, func(ctx context.Context) error {
		_, e := vpcClient.CreateSecurityGroupPolicies(pReq)
		return e
	})
	if err != nil {
		return Halt(state, err, "Failed to create securitygroup polices")
	}

	// bind securitygroup engress police
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
	err = Retry(ctx, func(ctx context.Context) error {
		_, e := vpcClient.CreateSecurityGroupPolicies(pReq)
		return e
	})
	if err != nil {
		return Halt(state, err, "Failed to create securitygroup polices")
	}

	Message(state, "Securitygroup polices created", "")

	return multistep.ActionContinue
}

func (s *stepConfigSecurityGroup) Cleanup(state multistep.StateBag) {
	if !s.isCreate {
		return
	}

	ctx := context.TODO()
	vpcClient := state.Get("vpc_client").(*vpc.Client)

	SayClean(state, "securitygroup")

	req := vpc.NewDeleteSecurityGroupRequest()
	req.SecurityGroupId = &s.SecurityGroupId
	err := Retry(ctx, func(ctx context.Context) error {
		_, e := vpcClient.DeleteSecurityGroup(req)
		return e
	})
	if err != nil {
		Error(state, err, fmt.Sprintf("Failed to delete securitygroup(%s), please delete it manually", s.SecurityGroupId))
	}
}
