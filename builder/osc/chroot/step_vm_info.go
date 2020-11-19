package chroot

import (
	"context"
	"fmt"
	"log"

	"github.com/antihax/optional"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/outscale/osc-sdk-go/osc"
)

// StepVmInfo verifies that this builder is running on an Outscale vm.
type StepVmInfo struct{}

func (s *StepVmInfo) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	oscconn := state.Get("osc").(*osc.APIClient)
	//session := state.Get("clientConfig").(*session.Session)
	ui := state.Get("ui").(packersdk.Ui)

	// Get our own vm ID
	ui.Say("Gathering information about this Outscale vm...")

	cmd := ShellCommand("curl http://169.254.169.254/latest/meta-data/instance-id")

	vmID, err := cmd.Output()
	if err != nil {
		err := fmt.Errorf(
			"Error retrieving the ID of the vm Packer is running on.\n" +
				"Please verify Packer is running on a proper Outscale vm.")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("[Debug] VmID got: %s", string(vmID))

	// Query the entire vm metadata
	resp, _, err := oscconn.VmApi.ReadVms(context.Background(), &osc.ReadVmsOpts{
		ReadVmsRequest: optional.NewInterface(osc.ReadVmsRequest{Filters: osc.FiltersVm{
			VmIds: []string{string(vmID)},
		}}),
	})
	if err != nil {
		err := fmt.Errorf("Error getting vm data: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	vmsResp := resp

	if len(vmsResp.Vms) == 0 {
		err := fmt.Errorf("Error getting vm data: no vm found")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	vm := vmsResp.Vms[0]
	state.Put("vm", vm)

	return multistep.ActionContinue
}

func (s *StepVmInfo) Cleanup(multistep.StateBag) {}
