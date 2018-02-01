package common

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepSecurityGroup struct {
	CommConfig            *communicator.Config
	SecurityGroupIds      []string
	VpcId                 string
	TemporarySGSourceCidr string

	createdGroupId string
}

func (s *StepSecurityGroup) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	if len(s.SecurityGroupIds) > 0 {
		_, err := ec2conn.DescribeSecurityGroups(
			&ec2.DescribeSecurityGroupsInput{
				GroupIds: aws.StringSlice(s.SecurityGroupIds),
			},
		)
		if err != nil {
			err := fmt.Errorf("Couldn't find specified security group: %s", err)
			log.Printf("[DEBUG] %s", err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
		log.Printf("Using specified security groups: %v", s.SecurityGroupIds)
		state.Put("securityGroupIds", s.SecurityGroupIds)
		return multistep.ActionContinue
	}

	port := s.CommConfig.Port()
	if port == 0 {
		if s.CommConfig.Type != "none" {
			panic("port must be set to a non-zero value.")
		}
	}

	// Create the group
	groupName := fmt.Sprintf("packer_%s", uuid.TimeOrderedUUID())
	ui.Say(fmt.Sprintf("Creating temporary security group for this instance: %s", groupName))
	group := &ec2.CreateSecurityGroupInput{
		GroupName:   &groupName,
		Description: aws.String("Temporary group for Packer"),
	}

	if s.VpcId != "" {
		group.VpcId = &s.VpcId
	}

	groupResp, err := ec2conn.CreateSecurityGroup(group)
	if err != nil {
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// Set the group ID so we can delete it later
	s.createdGroupId = *groupResp.GroupId

	// Authorize the SSH access for the security group
	req := &ec2.AuthorizeSecurityGroupIngressInput{
		GroupId:    groupResp.GroupId,
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int64(int64(port)),
		ToPort:     aws.Int64(int64(port)),
		CidrIp:     aws.String(s.TemporarySGSourceCidr),
	}

	// We loop and retry this a few times because sometimes the security
	// group isn't available immediately because AWS resources are eventually
	// consistent.
	ui.Say(fmt.Sprintf(
		"Authorizing access to port %d from %s in the temporary security group...",
		port, s.TemporarySGSourceCidr))
	for i := 0; i < 5; i++ {
		_, err = ec2conn.AuthorizeSecurityGroupIngress(req)
		if err == nil {
			break
		}

		log.Printf("Error authorizing. Will sleep and retry. %s", err)
		time.Sleep((time.Duration(i) * time.Second) + 1)
	}

	if err != nil {
		err := fmt.Errorf("Error creating temporary security group: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("[DEBUG] Waiting for temporary security group: %s", s.createdGroupId)
	err = waitUntilSecurityGroupExists(ec2conn,
		&ec2.DescribeSecurityGroupsInput{
			GroupIds: []*string{aws.String(s.createdGroupId)},
		},
	)
	if err == nil {
		log.Printf("[DEBUG] Found security group %s", s.createdGroupId)
	} else {
		err := fmt.Errorf("Timed out waiting for security group %s: %s", s.createdGroupId, err)
		log.Printf("[DEBUG] %s", err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// Set some state data for use in future steps
	state.Put("securityGroupIds", []string{s.createdGroupId})

	return multistep.ActionContinue
}

func (s *StepSecurityGroup) Cleanup(state multistep.StateBag) {
	if s.createdGroupId == "" {
		return
	}

	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting temporary security group...")

	var err error
	for i := 0; i < 5; i++ {
		_, err = ec2conn.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{GroupId: &s.createdGroupId})
		if err == nil {
			break
		}

		log.Printf("Error deleting security group: %s", err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up security group. Please delete the group manually: %s", s.createdGroupId))
	}
}

func waitUntilSecurityGroupExists(c *ec2.EC2, input *ec2.DescribeSecurityGroupsInput) error {
	ctx := aws.BackgroundContext()
	w := request.Waiter{
		Name:        "DescribeSecurityGroups",
		MaxAttempts: 40,
		Delay:       request.ConstantWaiterDelay(5 * time.Second),
		Acceptors: []request.WaiterAcceptor{
			{
				State:    request.SuccessWaiterState,
				Matcher:  request.PathWaiterMatch,
				Argument: "length(SecurityGroups[]) > `0`",
				Expected: true,
			},
			{
				State:    request.RetryWaiterState,
				Matcher:  request.ErrorWaiterMatch,
				Argument: "",
				Expected: "InvalidGroup.NotFound",
			},
			{
				State:    request.RetryWaiterState,
				Matcher:  request.ErrorWaiterMatch,
				Argument: "",
				Expected: "InvalidSecurityGroupID.NotFound",
			},
		},
		Logger: c.Config.Logger,
		NewRequest: func(opts []request.Option) (*request.Request, error) {
			var inCpy *ec2.DescribeSecurityGroupsInput
			if input != nil {
				tmp := *input
				inCpy = &tmp
			}
			req, _ := c.DescribeSecurityGroupsRequest(inCpy)
			req.SetContext(ctx)
			req.ApplyOptions(opts...)
			return req, nil
		},
	}
	return w.WaitWithContext(ctx)
}
