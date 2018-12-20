package cvm

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/packer/common"
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

func (s *stepConfigVPC) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
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

	vpcClient := state.Get("vpc_client").(*vpc.Client)
	ui := state.Get("ui").(packer.Ui)

	MessageClean(state, "VPC")
	req := vpc.NewDeleteVpcRequest()
	req.VpcId = &s.VpcId
	err := common.Retry(5, 5, 60, func(u uint) (bool, error) {
		_, err := vpcClient.DeleteVpc(req)
		if err == nil {
			return true, nil
		}
		if strings.Index(err.Error(), "ResourceInUse") != -1 {
			return false, nil
		} else {
			return false, err
		}
	})
	if err != nil {
		ui.Error(fmt.Sprintf("delete vpc(%s) failed: %s, you need to delete it by hand",
			s.VpcId, err.Error()))
		return
	}
}
