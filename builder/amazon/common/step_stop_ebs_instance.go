package common

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/builder/amazon/common/awserrors"
	"github.com/hashicorp/packer/common/retry"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepStopEBSBackedInstance struct {
	PollingConfig       *AWSPollingConfig
	Skip                bool
	DisableStopInstance bool
}

func (s *StepStopEBSBackedInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	instance := state.Get("instance").(*ec2.Instance)
	ui := state.Get("ui").(packer.Ui)

	// Skip when it is a spot instance
	if s.Skip {
		return multistep.ActionContinue
	}

	var err error

	if !s.DisableStopInstance {
		// Stop the instance so we can create an AMI from it
		ui.Say("Stopping the source instance...")

		// Amazon EC2 API follows an eventual consistency model.

		// This means that if you run a command to modify or describe a resource
		// that you just created, its ID might not have propagated throughout
		// the system, and you will get an error responding that the resource
		// does not exist.

		// Work around this by retrying a few times, up to about 5 minutes.
		err := retry.Config{Tries: 6, ShouldRetry: func(error) bool {
			if awserrors.Matches(err, "InvalidInstanceID.NotFound", "") {
				return true
			}
			return false
		},
			RetryDelay: (&retry.Backoff{InitialBackoff: 10 * time.Second, MaxBackoff: 60 * time.Second, Multiplier: 2}).Linear,
		}.Run(ctx, func(ctx context.Context) error {
			ui.Message(fmt.Sprintf("Stopping instance"))

			_, err = ec2conn.StopInstances(&ec2.StopInstancesInput{
				InstanceIds: []*string{instance.InstanceId},
			})

			return err
		})

		if err != nil {
			err := fmt.Errorf("Error stopping instance: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

	} else {
		ui.Say("Automatic instance stop disabled. Please stop instance manually.")
	}

	// Wait for the instance to actually stop
	ui.Say("Waiting for the instance to stop...")
	err = ec2conn.WaitUntilInstanceStoppedWithContext(ctx,
		&ec2.DescribeInstancesInput{
			InstanceIds: []*string{instance.InstanceId},
		},
		s.PollingConfig.getWaiterOptions()...)

	if err != nil {
		err := fmt.Errorf("Error waiting for instance to stop: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepStopEBSBackedInstance) Cleanup(multistep.StateBag) {
	// No cleanup...
}
