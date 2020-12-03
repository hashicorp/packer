package ecs

import (
	"context"
	errorsNew "errors"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/uuid"
)

type stepConfigAlicloudVPC struct {
	VpcId     string
	CidrBlock string //192.168.0.0/16 or 172.16.0.0/16 (default)
	VpcName   string
	isCreate  bool
}

var createVpcRetryErrors = []string{
	"TOKEN_PROCESSING",
}

var deleteVpcRetryErrors = []string{
	"DependencyViolation.Instance",
	"DependencyViolation.RouteEntry",
	"DependencyViolation.VSwitch",
	"DependencyViolation.SecurityGroup",
	"Forbbiden",
	"TaskConflict",
}

func (s *stepConfigAlicloudVPC) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packersdk.Ui)

	if len(s.VpcId) != 0 {
		describeVpcsRequest := ecs.CreateDescribeVpcsRequest()
		describeVpcsRequest.VpcId = s.VpcId
		describeVpcsRequest.RegionId = config.AlicloudRegion

		vpcsResponse, err := client.DescribeVpcs(describeVpcsRequest)
		if err != nil {
			return halt(state, err, "Failed querying vpcs")
		}

		vpcs := vpcsResponse.Vpcs.Vpc
		if len(vpcs) > 0 {
			state.Put("vpcid", vpcs[0].VpcId)
			s.isCreate = false
			return multistep.ActionContinue
		}

		message := fmt.Sprintf("The specified vpc {%s} doesn't exist.", s.VpcId)
		return halt(state, errorsNew.New(message), "")
	}

	ui.Say("Creating vpc...")

	createVpcRequest := s.buildCreateVpcRequest(state)
	createVpcResponse, err := client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			return client.CreateVpc(createVpcRequest)
		},
		EvalFunc: client.EvalCouldRetryResponse(createVpcRetryErrors, EvalRetryErrorType),
	})
	if err != nil {
		return halt(state, err, "Failed creating vpc")
	}

	vpcId := createVpcResponse.(*ecs.CreateVpcResponse).VpcId
	_, err = client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			request := ecs.CreateDescribeVpcsRequest()
			request.RegionId = config.AlicloudRegion
			request.VpcId = vpcId
			return client.DescribeVpcs(request)
		},
		EvalFunc: func(response responses.AcsResponse, err error) WaitForExpectEvalResult {
			if err != nil {
				return WaitForExpectToRetry
			}

			vpcsResponse := response.(*ecs.DescribeVpcsResponse)
			vpcs := vpcsResponse.Vpcs.Vpc
			if len(vpcs) > 0 {
				for _, vpc := range vpcs {
					if vpc.Status == VpcStatusAvailable {
						return WaitForExpectSuccess
					}
				}
			}

			return WaitForExpectToRetry
		},
		RetryTimes: shortRetryTimes,
	})

	if err != nil {
		return halt(state, err, "Failed waiting for vpc to become available")
	}

	ui.Message(fmt.Sprintf("Created vpc: %s", vpcId))
	state.Put("vpcid", vpcId)
	s.isCreate = true
	s.VpcId = vpcId
	return multistep.ActionContinue
}

func (s *stepConfigAlicloudVPC) Cleanup(state multistep.StateBag) {
	if !s.isCreate {
		return
	}

	cleanUpMessage(state, "VPC")

	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packersdk.Ui)

	_, err := client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			request := ecs.CreateDeleteVpcRequest()
			request.VpcId = s.VpcId
			return client.DeleteVpc(request)
		},
		EvalFunc:   client.EvalCouldRetryResponse(deleteVpcRetryErrors, EvalRetryErrorType),
		RetryTimes: shortRetryTimes,
	})

	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting vpc, it may still be around: %s", err))
	}
}

func (s *stepConfigAlicloudVPC) buildCreateVpcRequest(state multistep.StateBag) *ecs.CreateVpcRequest {
	config := state.Get("config").(*Config)

	request := ecs.CreateCreateVpcRequest()
	request.ClientToken = uuid.TimeOrderedUUID()
	request.RegionId = config.AlicloudRegion
	request.CidrBlock = s.CidrBlock
	request.VpcName = s.VpcName

	return request
}
