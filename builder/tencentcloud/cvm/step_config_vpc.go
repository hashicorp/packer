package cvm

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/common/retry"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/pkg/errors"
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
	ui := state.Get("ui").(packer.Ui)

	if len(s.VpcId) != 0 { // exist vpc
		ui.Say(fmt.Sprintf("Trying to use existing vpc(%s)", s.VpcId))
		req := vpc.NewDescribeVpcsRequest()
		req.VpcIds = []*string{&s.VpcId}
		resp, err := vpcClient.DescribeVpcs(req)
		if err != nil {
			ui.Error(fmt.Sprintf("query vpc failed: %s", err.Error()))
			state.Put("error", err)
			return multistep.ActionHalt
		}
		if *resp.Response.TotalCount > 0 {
			vpc0 := *resp.Response.VpcSet[0]
			state.Put("vpc_id", *vpc0.VpcId)
			s.isCreate = false
			return multistep.ActionContinue
		}
		message := fmt.Sprintf("the specified vpc(%s) does not exist", s.VpcId)
		state.Put("error", errors.New(message))
		ui.Error(message)
		return multistep.ActionHalt
	} else { // create a new vpc, tencentcloud create vpc api is synchronous, no need to wait for create.
		ui.Say(fmt.Sprintf("Trying to create a new vpc"))
		req := vpc.NewCreateVpcRequest()
		req.VpcName = &s.VpcName
		req.CidrBlock = &s.CidrBlock
		resp, err := vpcClient.CreateVpc(req)
		if err != nil {
			ui.Error(fmt.Sprintf("create vpc failed: %s", err.Error()))
			state.Put("error", err)
			return multistep.ActionHalt
		}
		vpc0 := *resp.Response.Vpc
		state.Put("vpc_id", *vpc0.VpcId)
		s.VpcId = *vpc0.VpcId
		s.isCreate = true
		return multistep.ActionContinue
	}
}

func (s *stepConfigVPC) Cleanup(state multistep.StateBag) {
	if !s.isCreate {
		return
	}
	ctx := context.TODO()

	vpcClient := state.Get("vpc_client").(*vpc.Client)
	ui := state.Get("ui").(packer.Ui)

	MessageClean(state, "VPC")
	req := vpc.NewDeleteVpcRequest()
	req.VpcId = &s.VpcId
	err := retry.Config{
		Tries:      60,
		RetryDelay: (&retry.Backoff{InitialBackoff: 5 * time.Second, MaxBackoff: 5 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		_, err := vpcClient.DeleteVpc(req)
		return err
	})
	if err != nil {
		ui.Error(fmt.Sprintf("delete vpc(%s) failed: %s, you need to delete it by hand",
			s.VpcId, err.Error()))
		return
	}
}
