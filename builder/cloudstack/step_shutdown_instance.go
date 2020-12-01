package cloudstack

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

type stepShutdownInstance struct{}

func (s *stepShutdownInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*cloudstack.CloudStackClient)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Shutting down instance...")

	// Retrieve the instance ID from the previously saved state.
	instanceID, ok := state.Get("instance_id").(string)
	if !ok || instanceID == "" {
		state.Put("error", fmt.Errorf("Could not retrieve instance_id from state!"))
		return multistep.ActionHalt
	}

	// Create a new parameter struct.
	p := client.VirtualMachine.NewStopVirtualMachineParams(instanceID)

	// Shutdown the virtual machine.
	_, err := client.VirtualMachine.StopVirtualMachine(p)
	if err != nil {
		err := fmt.Errorf("Error shutting down instance %s: %s", config.InstanceName, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message("Instance has been shutdown!")
	return multistep.ActionContinue
}

// Cleanup any resources that may have been created during the Run phase.
func (s *stepShutdownInstance) Cleanup(state multistep.StateBag) {
	// Nothing to cleanup for this step.
}
