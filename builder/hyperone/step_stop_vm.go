package hyperone

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	openapi "github.com/hyperonecom/h1-client-go"
)

type stepStopVM struct{}

func (s *stepStopVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*openapi.APIClient)
	ui := state.Get("ui").(packersdk.Ui)
	vmID := state.Get("vm_id").(string)

	ui.Say("Stopping VM...")

	_, _, err := client.VmApi.VmActionStop(ctx, vmID)
	if err != nil {
		err := fmt.Errorf("error stopping VM: %s", formatOpenAPIError(err))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepStopVM) Cleanup(multistep.StateBag) {}
