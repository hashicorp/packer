package googlecompute

import (
	"fmt"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// stepInstanceInfo represents a Packer build step that gathers GCE instance info.
type stepInstanceInfo int

// Run executes the Packer build step that gathers GCE instance info.
func (s *stepInstanceInfo) Run(state multistep.StateBag) multistep.StepAction {
	var (
		client = state.Get("client").(*GoogleComputeClient)
		config = state.Get("config").(*Config)
		ui     = state.Get("ui").(packer.Ui)
	)
	instanceName := state.Get("instance_name").(string)
	err := waitForInstanceState("RUNNING", config.Zone, instanceName, client, config.stateTimeout)
	if err != nil {
		err := fmt.Errorf("Error creating instance: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	ip, err := client.GetNatIP(config.Zone, instanceName)
	if err != nil {
		err := fmt.Errorf("Error retrieving instance nat ip address: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("instance_ip", ip)
	return multistep.ActionContinue
}

// Cleanup.
func (s *stepInstanceInfo) Cleanup(state multistep.StateBag) {}
