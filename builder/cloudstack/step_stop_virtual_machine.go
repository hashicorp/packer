package cloudstack

import (
	"fmt"
	"github.com/mindjiver/gopherstack"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

type stepStopVirtualMachine struct{}

func (s *stepStopVirtualMachine) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*gopherstack.CloudstackClient)
	c := state.Get("config").(config)
	ui := state.Get("ui").(packer.Ui)
	id := state.Get("virtual_machine_id").(string)

	response, err := client.ListVirtualMachines(id)
	if err != nil {
		err := fmt.Errorf("Error checking virtual machine state: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// As we list the virtual machines with the unique UUID we
	// know the VM we are after is the first one.
	currentState := response.Listvirtualmachinesresponse.Virtualmachine[0].State
	if currentState == "Stopped" {
		// Virtual Machine is already stopped, don't do anything
		return multistep.ActionContinue
	}

	// Stop the virtual machine
	ui.Say("Stopping virtual machine...")
	response2, err := client.StopVirtualMachine(id)
	if err != nil {
		err := fmt.Errorf("Error stopping virtual machine: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Println("Waiting for stop event to complete...")
	jobid := response2.Stopvirtualmachineresponse.Jobid
	err = client.WaitForAsyncJob(jobid, c.stateTimeout)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepStopVirtualMachine) Cleanup(state multistep.StateBag) {
	// no cleanup
}
