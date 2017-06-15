package cloudstack

import (
	"fmt"
	"time"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

type stepDetachIso struct {
	DetachISO     bool
	DetachISOWait time.Duration
}

func (s *stepDetachIso) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*cloudstack.CloudStackClient)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	if !config.DetachISO {
		return multistep.ActionContinue
	}

	instanceId := state.Get("instance_id").(string)
	params := client.VirtualMachine.NewListVirtualMachinesParams()
	params.SetId(instanceId)
	vmId, err := client.VirtualMachine.ListVirtualMachines(params)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	isoId := vmId.VirtualMachines[0].Isoid
	if isoId == "" {
		// No ISO image attached
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Waiting for %v before detaching ISO from virtual machine...", config.DetachISOWait))
	time.Sleep(config.DetachISOWait)
	ui.Say("Detaching ISO...")
	_, err = client.ISO.DetachIso(client.ISO.NewDetachIsoParams(instanceId))
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepDetachIso) Cleanup(state multistep.StateBag) {}
