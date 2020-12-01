package oci

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepCreateInstance struct{}

func (s *stepCreateInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	var (
		driver = state.Get("driver").(Driver)
		ui     = state.Get("ui").(packersdk.Ui)
		config = state.Get("config").(*Config)
	)

	ui.Say("Creating instance...")

	instanceID, err := driver.CreateInstance(ctx, string(config.Comm.SSHPublicKey))
	if err != nil {
		err = fmt.Errorf("Problem creating instance: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	state.Put("instance_id", instanceID)

	ui.Say(fmt.Sprintf("Created instance (%s).", instanceID))

	ui.Say("Waiting for instance to enter 'RUNNING' state...")

	if err = driver.WaitForInstanceState(ctx, instanceID, []string{"STARTING", "PROVISIONING"}, "RUNNING"); err != nil {
		err = fmt.Errorf("Error waiting for instance to start: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say("Instance 'RUNNING'.")

	return multistep.ActionContinue
}

func (s *stepCreateInstance) Cleanup(state multistep.StateBag) {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packersdk.Ui)

	idRaw, ok := state.GetOk("instance_id")
	if !ok {
		return
	}
	id := idRaw.(string)

	ui.Say(fmt.Sprintf("Terminating instance (%s)...", id))

	if err := driver.TerminateInstance(context.TODO(), id); err != nil {
		err = fmt.Errorf("Error terminating instance. Please terminate manually: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}

	err := driver.WaitForInstanceState(context.TODO(), id, []string{"TERMINATING"}, "TERMINATED")
	if err != nil {
		err = fmt.Errorf("Error terminating instance. Please terminate manually: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return
	}

	ui.Say("Terminated instance.")
}
