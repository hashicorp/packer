package cvm

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

type stepConfigSubnet struct {
	SubnetId        string
	SubnetCidrBlock string
	SubnetName      string
	Zone            string
	isCreate        bool
}

func (s *stepConfigSubnet) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	vpcClient := state.Get("vpc_client").(*vpc.Client)

	vpcId := state.Get("vpc_id").(string)

	if len(s.SubnetId) != 0 {
		Say(state, s.SubnetId, "Trying to use existing subnet")
		req := vpc.NewDescribeSubnetsRequest()
		req.SubnetIds = []*string{&s.SubnetId}
		var resp *vpc.DescribeSubnetsResponse
		err := Retry(ctx, func(ctx context.Context) error {
			var e error
			resp, e = vpcClient.DescribeSubnets(req)
			return e
		})
		if err != nil {
			return Halt(state, err, "Failed to get subnet info")
		}
		if *resp.Response.TotalCount > 0 {
			s.isCreate = false
			if *resp.Response.SubnetSet[0].VpcId != vpcId {
				return Halt(state, fmt.Errorf("The specified subnet(%s) does not belong to the specified vpc(%s)", s.SubnetId, vpcId), "")
			}
			state.Put("subnet_id", *resp.Response.SubnetSet[0].SubnetId)
			Message(state, *resp.Response.SubnetSet[0].SubnetName, "Subnet found")
			return multistep.ActionContinue
		}
		return Halt(state, fmt.Errorf("The specified subnet(%s) does not exist", s.SubnetId), "")
	}

	Say(state, "Trying to create a new subnet", "")

	req := vpc.NewCreateSubnetRequest()
	req.VpcId = &vpcId
	req.SubnetName = &s.SubnetName
	req.CidrBlock = &s.SubnetCidrBlock
	req.Zone = &s.Zone
	var resp *vpc.CreateSubnetResponse
	err := Retry(ctx, func(ctx context.Context) error {
		var e error
		resp, e = vpcClient.CreateSubnet(req)
		return e
	})
	if err != nil {
		return Halt(state, err, "Failed to create subnet")
	}

	s.isCreate = true
	s.SubnetId = *resp.Response.Subnet.SubnetId
	state.Put("subnet_id", s.SubnetId)
	Message(state, s.SubnetId, "Subnet created")

	return multistep.ActionContinue
}

func (s *stepConfigSubnet) Cleanup(state multistep.StateBag) {
	if !s.isCreate {
		return
	}

	ctx := context.TODO()
	vpcClient := state.Get("vpc_client").(*vpc.Client)

	SayClean(state, "subnet")

	req := vpc.NewDeleteSubnetRequest()
	req.SubnetId = &s.SubnetId
	err := Retry(ctx, func(ctx context.Context) error {
		_, e := vpcClient.DeleteSubnet(req)
		return e
	})
	if err != nil {
		Error(state, err, fmt.Sprintf("Failed to delete subnet(%s), please delete it manually", s.SubnetId))
	}
}
