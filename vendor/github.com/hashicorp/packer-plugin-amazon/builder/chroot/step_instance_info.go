package chroot

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// StepInstanceInfo verifies that this builder is running on an EC2 instance.
type StepInstanceInfo struct{}

func (s *StepInstanceInfo) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	session := state.Get("awsSession").(*session.Session)
	ui := state.Get("ui").(packersdk.Ui)

	// Get our own instance ID
	ui.Say("Gathering information about this EC2 instance...")

	ec2meta := ec2metadata.New(session)
	identity, err := ec2meta.GetInstanceIdentityDocument()
	if err != nil {
		err := fmt.Errorf(
			"Error retrieving the ID of the instance Packer is running on.\n" +
				"Please verify Packer is running on a proper AWS EC2 instance.")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	log.Printf("Instance ID: %s", identity.InstanceID)

	// Query the entire instance metadata
	instancesResp, err := ec2conn.DescribeInstances(&ec2.DescribeInstancesInput{InstanceIds: []*string{&identity.InstanceID}})
	if err != nil {
		err := fmt.Errorf("Error getting instance data: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(instancesResp.Reservations) == 0 {
		err := fmt.Errorf("Error getting instance data: no instance found.")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	instance := instancesResp.Reservations[0].Instances[0]
	state.Put("instance", instance)

	return multistep.ActionContinue
}

func (s *StepInstanceInfo) Cleanup(multistep.StateBag) {}
