package common

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepStopEBSBackedInstance struct {
	SpotPrice           string
	DisableStopInstance bool
}

func (s *StepStopEBSBackedInstance) Run(state multistep.StateBag) multistep.StepAction {
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
	stateChange := StateChangeConf{
		Pending:   []string{"running", "stopping"},
		Target:    "stopped",
		Refresh:   InstanceStateRefreshFunc(ec2conn, *instance.InstanceId),
		StepState: state,
	}
	_, err = WaitForState(&stateChange)
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
