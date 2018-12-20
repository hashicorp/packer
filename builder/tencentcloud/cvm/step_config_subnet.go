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

type stepConfigSubnet struct {
	SubnetId        string
	SubnetCidrBlock string
	SubnetName      string
	Zone            string
	isCreate        bool
}

func (s *stepConfigSubnet) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	vpcClient := state.Get("vpc_client").(*vpc.Client)
	ui := state.Get("ui").(packer.Ui)
	vpcId := state.Get("vpc_id").(string)

	if len(s.SubnetId) != 0 { // exist subnet
		ui.Say(fmt.Sprintf("Trying to use existing subnet(%s)", s.SubnetId))
		req := vpc.NewDescribeSubnetsRequest()
		req.SubnetIds = []*string{&s.SubnetId}
		resp, err := vpcClient.DescribeSubnets(req)
		if err != nil {
			ui.Error(fmt.Sprintf("query subnet failed: %s", err.Error()))
			state.Put("error", err)
			return multistep.ActionHalt
		}
		if *resp.Response.TotalCount > 0 {
			subnet0 := *resp.Response.SubnetSet[0]
			if *subnet0.VpcId != vpcId {
				message := fmt.Sprintf("the specified subnet(%s) does not belong to "+
					"the specified vpc(%s)", s.SubnetId, vpcId)
				ui.Error(message)
				state.Put("error", errors.New(message))
				return multistep.ActionHalt
			}
			state.Put("subnet_id", *subnet0.SubnetId)
			s.isCreate = false
			return multistep.ActionContinue
		}
		message := fmt.Sprintf("the specified subnet(%s) does not exist", s.SubnetId)
		state.Put("error", errors.New(message))
		ui.Error(message)
		return multistep.ActionHalt
	} else { // create a new subnet, tencentcloud create subnet api is synchronous, no need to wait for create.
		ui.Say(fmt.Sprintf("Trying to create a new subnet"))
		req := vpc.NewCreateSubnetRequest()
		req.VpcId = &vpcId
		req.SubnetName = &s.SubnetName
		req.CidrBlock = &s.SubnetCidrBlock
		req.Zone = &s.Zone
		resp, err := vpcClient.CreateSubnet(req)
		if err != nil {
			ui.Error(fmt.Sprintf("create subnet failed: %s", err.Error()))
			state.Put("error", err)
			return multistep.ActionHalt
		}
		subnet0 := *resp.Response.Subnet
		state.Put("subnet_id", *subnet0.SubnetId)
		s.SubnetId = *subnet0.SubnetId
		s.isCreate = true
		return multistep.ActionContinue
	}
}

func (s *stepConfigSubnet) Cleanup(state multistep.StateBag) {
	if !s.isCreate {
		return
	}

	vpcClient := state.Get("vpc_client").(*vpc.Client)
	ui := state.Get("ui").(packer.Ui)

	MessageClean(state, "SUBNET")
	req := vpc.NewDeleteSubnetRequest()
	req.SubnetId = &s.SubnetId
	err := common.Retry(5, 5, 60, func(u uint) (bool, error) {
		_, err := vpcClient.DeleteSubnet(req)
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
		ui.Error(fmt.Sprintf("delete subnet(%s) failed: %s, you need to delete it by hand",
			s.SubnetId, err.Error()))
		return
	}
}
