package classic

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepTerminatePVMaster struct {
}

func (s *stepTerminatePVMaster) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	// get variables from state
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Deleting master Instance...")

	client := state.Get("client").(*compute.Client)
	instanceInfo := state.Get("master_instance_info").(*compute.InstanceInfo)

	// get instances client
	instanceClient := client.Instances()

	// Instances Input
	input := &compute.DeleteInstanceInput{
		Name: instanceInfo.Name,
		ID:   instanceInfo.ID,
	}

	err := instanceClient.DeleteInstance(input)
	if err != nil {
		err = fmt.Errorf("Problem creating instance: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Deleted master instance: %s.", instanceInfo.ID))
	state.Put("master_instance_deleted", true)
	return multistep.ActionContinue
}

func (s *stepTerminatePVMaster) Cleanup(state multistep.StateBag) {
}
