package cvm

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

type stepConfigVPC struct {
	VpcId     string
	CidrBlock string
	VpcName   string
	isCreate  bool
}

func (s *stepConfigVPC) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	vpcClient := state.Get("vpc_client").(*vpc.Client)

	if len(s.VpcId) != 0 {
		Say(state, s.VpcId, "Trying to use existing vpc")
		req := vpc.NewDescribeVpcsRequest()
		req.VpcIds = []*string{&s.VpcId}
		var resp *vpc.DescribeVpcsResponse
		err := Retry(ctx, func(ctx context.Context) error {
			var e error
			resp, e = vpcClient.DescribeVpcs(req)
			return e
		})
		if err != nil {
			return Halt(state, err, "Failed to get vpc info")
		}
		if *resp.Response.TotalCount > 0 {
			s.isCreate = false
			state.Put("vpc_id", *resp.Response.VpcSet[0].VpcId)
			Message(state, *resp.Response.VpcSet[0].VpcName, "Vpc found")
			return multistep.ActionContinue
		}
		return Halt(state, fmt.Errorf("The specified vpc(%s) does not exist", s.VpcId), "")
	}

	Say(state, "Trying to create a new vpc", "")

	req := vpc.NewCreateVpcRequest()
	req.VpcName = &s.VpcName
	req.CidrBlock = &s.CidrBlock
	var resp *vpc.CreateVpcResponse
	err := Retry(ctx, func(ctx context.Context) error {
		var e error
		resp, e = vpcClient.CreateVpc(req)
		return e
	})
	if err != nil {
		return Halt(state, err, "Failed to create vpc")
	}

	s.isCreate = true
	s.VpcId = *resp.Response.Vpc.VpcId
	state.Put("vpc_id", s.VpcId)
	Message(state, s.VpcId, "Vpc created")

	return multistep.ActionContinue
}

func (s *stepConfigVPC) Cleanup(state multistep.StateBag) {
	if !s.isCreate {
		return
	}

	ctx := context.TODO()
	vpcClient := state.Get("vpc_client").(*vpc.Client)

	SayClean(state, "vpc")

	req := vpc.NewDeleteVpcRequest()
	req.VpcId = &s.VpcId
	err := Retry(ctx, func(ctx context.Context) error {
		_, e := vpcClient.DeleteVpc(req)
		return e
	})
	if err != nil {
		Error(state, err, fmt.Sprintf("Failed to delete vpc(%s), please delete it manually", s.VpcId))
	}
}
