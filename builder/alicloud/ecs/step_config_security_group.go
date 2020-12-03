package ecs

import (
	"context"
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/uuid"
)

type stepConfigAlicloudSecurityGroup struct {
	SecurityGroupId   string
	SecurityGroupName string
	Description       string
	VpcId             string
	RegionId          string
	isCreate          bool
}

var createSecurityGroupRetryErrors = []string{
	"IdempotentProcessing",
}

var deleteSecurityGroupRetryErrors = []string{
	"DependencyViolation",
}

func (s *stepConfigAlicloudSecurityGroup) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packersdk.Ui)
	networkType := state.Get("networktype").(InstanceNetWork)

	if len(s.SecurityGroupId) != 0 {
		describeSecurityGroupsRequest := ecs.CreateDescribeSecurityGroupsRequest()
		describeSecurityGroupsRequest.RegionId = s.RegionId
		describeSecurityGroupsRequest.SecurityGroupId = s.SecurityGroupId
		if networkType == InstanceNetworkVpc {
			vpcId := state.Get("vpcid").(string)
			describeSecurityGroupsRequest.VpcId = vpcId
		}

		securityGroupsResponse, err := client.DescribeSecurityGroups(describeSecurityGroupsRequest)
		if err != nil {
			return halt(state, err, "Failed querying security group")
		}

		securityGroupItems := securityGroupsResponse.SecurityGroups.SecurityGroup
		for _, securityGroupItem := range securityGroupItems {
			if securityGroupItem.SecurityGroupId == s.SecurityGroupId {
				state.Put("securitygroupid", s.SecurityGroupId)
				s.isCreate = false
				return multistep.ActionContinue
			}
		}

		s.isCreate = false
		err = fmt.Errorf("The specified security group {%s} doesn't exist.", s.SecurityGroupId)
		return halt(state, err, "")
	}

	ui.Say("Creating security group...")

	createSecurityGroupRequest := s.buildCreateSecurityGroupRequest(state)
	securityGroupResponse, err := client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			return client.CreateSecurityGroup(createSecurityGroupRequest)
		},
		EvalFunc: client.EvalCouldRetryResponse(createSecurityGroupRetryErrors, EvalRetryErrorType),
	})

	if err != nil {
		return halt(state, err, "Failed creating security group")
	}

	securityGroupId := securityGroupResponse.(*ecs.CreateSecurityGroupResponse).SecurityGroupId

	ui.Message(fmt.Sprintf("Created security group: %s", securityGroupId))
	state.Put("securitygroupid", securityGroupId)
	s.isCreate = true
	s.SecurityGroupId = securityGroupId

	authorizeSecurityGroupEgressRequest := ecs.CreateAuthorizeSecurityGroupEgressRequest()
	authorizeSecurityGroupEgressRequest.SecurityGroupId = securityGroupId
	authorizeSecurityGroupEgressRequest.RegionId = s.RegionId
	authorizeSecurityGroupEgressRequest.IpProtocol = IpProtocolAll
	authorizeSecurityGroupEgressRequest.PortRange = DefaultPortRange
	authorizeSecurityGroupEgressRequest.NicType = NicTypeInternet
	authorizeSecurityGroupEgressRequest.DestCidrIp = DefaultCidrIp

	if _, err := client.AuthorizeSecurityGroupEgress(authorizeSecurityGroupEgressRequest); err != nil {
		return halt(state, err, "Failed authorizing security group")
	}

	authorizeSecurityGroupRequest := ecs.CreateAuthorizeSecurityGroupRequest()
	authorizeSecurityGroupRequest.SecurityGroupId = securityGroupId
	authorizeSecurityGroupRequest.RegionId = s.RegionId
	authorizeSecurityGroupRequest.IpProtocol = IpProtocolAll
	authorizeSecurityGroupRequest.PortRange = DefaultPortRange
	authorizeSecurityGroupRequest.NicType = NicTypeInternet
	authorizeSecurityGroupRequest.SourceCidrIp = DefaultCidrIp

	if _, err := client.AuthorizeSecurityGroup(authorizeSecurityGroupRequest); err != nil {
		return halt(state, err, "Failed authorizing security group")
	}

	return multistep.ActionContinue
}

func (s *stepConfigAlicloudSecurityGroup) Cleanup(state multistep.StateBag) {
	if !s.isCreate {
		return
	}

	cleanUpMessage(state, "security group")

	client := state.Get("client").(*ClientWrapper)
	ui := state.Get("ui").(packersdk.Ui)

	_, err := client.WaitForExpected(&WaitForExpectArgs{
		RequestFunc: func() (responses.AcsResponse, error) {
			request := ecs.CreateDeleteSecurityGroupRequest()
			request.RegionId = s.RegionId
			request.SecurityGroupId = s.SecurityGroupId
			return client.DeleteSecurityGroup(request)
		},
		EvalFunc:   client.EvalCouldRetryResponse(deleteSecurityGroupRetryErrors, EvalRetryErrorType),
		RetryTimes: shortRetryTimes,
	})

	if err != nil {
		ui.Error(fmt.Sprintf("Failed to delete security group, it may still be around: %s", err))
	}
}

func (s *stepConfigAlicloudSecurityGroup) buildCreateSecurityGroupRequest(state multistep.StateBag) *ecs.CreateSecurityGroupRequest {
	networkType := state.Get("networktype").(InstanceNetWork)

	request := ecs.CreateCreateSecurityGroupRequest()
	request.ClientToken = uuid.TimeOrderedUUID()
	request.RegionId = s.RegionId
	request.SecurityGroupName = s.SecurityGroupName

	if networkType == InstanceNetworkVpc {
		vpcId := state.Get("vpcid").(string)
		request.VpcId = vpcId
	}

	return request
}
