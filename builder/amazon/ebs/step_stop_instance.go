package ebs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	awscommon "github.com/mitchellh/packer/builder/amazon/common"
	"github.com/mitchellh/packer/packer"
)

type stepStopInstance struct {
	SpotPrice           string
	DisableStopInstance bool
}

func (s *stepStopInstance) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	instance := state.Get("instance").(*ec2.Instance)
	ui := state.Get("ui").(packer.Ui)

	// Skip when it is a spot instance
	if s.SpotPrice != "" && s.SpotPrice != "0" {
		return multistep.ActionContinue
	}

	var err error

	if !s.DisableStopInstance {
		// Stop the instance so we can create an AMI from it
		ui.Say("Stopping the source instance...")
		_, err = ec2conn.StopInstances(&ec2.StopInstancesInput{
			InstanceIds: []*string{instance.InstanceId},
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

	// Wait for the instance to actual stop
	ui.Say("Waiting for the instance to stop...")
	stateChange := awscommon.StateChangeConf{
		Pending:   []string{"running", "stopping"},
		Target:    "stopped",
		Refresh:   awscommon.InstanceStateRefreshFunc(ec2conn, *instance.InstanceId),
		StepState: state,
	}
	_, err = awscommon.WaitForState(&stateChange)
	if err != nil {
		err := fmt.Errorf("Error waiting for instance to stop: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepStopInstance) Cleanup(multistep.StateBag) {
	// No cleanup...
}
