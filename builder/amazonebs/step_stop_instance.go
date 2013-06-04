package amazonebs

import (
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepStopInstance struct{}

func (s *stepStopInstance) Run(state map[string]interface{}) multistep.StepAction {
	ec2conn := state["ec2"].(*ec2.EC2)
	instance := state["instance"].(*ec2.Instance)
	ui := state["ui"].(packer.Ui)

	// Stop the instance so we can create an AMI from it
	ui.Say("Stopping the source instance...")
	_, err := ec2conn.StopInstances(instance.InstanceId)
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Wait for the instance to actual stop
	// TODO(mitchellh): Handle diff source states, i.e. this force state sucks
	ui.Say("Waiting for the instance to stop...")
	instance.State.Name = "stopping"
	instance, err = waitForState(ec2conn, instance, []string{"running", "stopping"}, "stopped")
	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepStopInstance) Cleanup(map[string]interface{}) {
	// No cleanup...
}
