package instance

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepBundleVolume struct{}

func (s *StepBundleVolume) Run(state map[string]interface{}) multistep.StepAction {
	comm := state["communicator"].(packer.Communicator)
	ui := state["ui"].(packer.Ui)

	// Verify the AMI tools are available
	ui.Say("Checking for EC2 AMI tools...")
	cmd := &packer.RemoteCmd{Command: "ec2-ami-tools-version"}
	if err := comm.Start(cmd); err != nil {
		state["error"] = fmt.Errorf("Error checking for AMI tools: %s", err)
		ui.Error(state["error"].(error).Error())
		return multistep.ActionHalt
	}
	cmd.Wait()

	if cmd.ExitStatus != 0 {
		state["error"] = fmt.Errorf(
			"The EC2 AMI tools could not be detected. These must be manually\n" +
				"via a provisioner or some other means and are required for Packer\n" +
				"to create an instance-store AMI.")
		ui.Error(state["error"].(error).Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepBundleVolume) Cleanup(map[string]interface{}) {}
