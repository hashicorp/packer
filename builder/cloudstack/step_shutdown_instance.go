package cloudstack

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

type stepShutdownInstance struct{}

func (s *stepShutdownInstance) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*cloudstack.CloudStackClient)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Shutting down instance...")

	// Retrieve the instance ID from the previously saved state.
	instanceID, ok := state.Get("instance_id").(string)
	if !ok || instanceID == "" {
		ui.Error("Could not retrieve instance_id from state!")
		return multistep.ActionHalt
	}

	// Create a new parameter struct.
	p := client.VirtualMachine.NewStopVirtualMachineParams(instanceID)

	// Shutdown the virtual machine.
	_, err := client.VirtualMachine.StopVirtualMachine(p)
	if err != nil {
		ui.Error(fmt.Sprintf("Error shutting down instance %s: %s", config.InstanceName, err))
		return multistep.ActionHalt
	}

	ui.Message("Instance has been shutdown!")

	return multistep.ActionContinue
}

// Cleanup any resources that may have been created during the Run phase.
func (s *stepShutdownInstance) Cleanup(state multistep.StateBag) {
	// Nothing to cleanup for this step.
}
