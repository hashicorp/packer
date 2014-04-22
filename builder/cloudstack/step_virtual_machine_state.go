package cloudstack

import (
	"fmt"
	"github.com/mindjiver/gopherstack"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"time"
)

type stepVirtualMachineState struct{}

func (s *stepVirtualMachineState) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*gopherstack.CloudstackClient)
	ui := state.Get("ui").(packer.Ui)
	id := state.Get("virtual_machine_id").(string)

	ui.Say("Waiting for virtual machine to become active...")

	err := client.WaitForVirtualMachineState(id, "Running", 2*time.Minute)
	if err != nil {
		err := fmt.Errorf("Error waiting for virtual machine to become active: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the IP on the state for later
	response, err := client.ListVirtualMachines(id)
	if err != nil {
		err := fmt.Errorf("Error retrieving virtual machine ID: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ip := response.Listvirtualmachinesresponse.Virtualmachine[0].Nic[0].Ipaddress
	state.Put("virtual_machine_ip", ip)

	return multistep.ActionContinue
}

func (s *stepVirtualMachineState) Cleanup(state multistep.StateBag) {
	// no cleanup
}
