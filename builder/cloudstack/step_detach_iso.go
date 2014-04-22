package cloudstack

import (
	"fmt"
	"github.com/mindjiver/gopherstack"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"time"
)

type stepDetachIso struct{}

func (s *stepDetachIso) Run(state multistep.StateBag) multistep.StepAction {
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
	isoId := response.Listvirtualmachinesresponse.Virtualmachine[0].IsoId
	if isoId == "" {
		// No ISO image attached to virtual machine, we just
		// continue
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Waiting for %v before detaching ISO from virtual machine...", c.detachISOWait))
	time.Sleep(c.detachISOWait)

	response2, err := client.DetachIso(id)
	if err != nil {
		err := fmt.Errorf("Error detaching ISO from virtual machine: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Println("Waiting for detach event to complete...")
	jobid := response2.Detachisoresponse.Jobid
	err = client.WaitForAsyncJob(jobid, c.stateTimeout)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepDetachIso) Cleanup(state multistep.StateBag) {
	// no cleanup
}
