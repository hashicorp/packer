package hyperone

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	openapi "github.com/hyperonecom/h1-client-go"
)

type stepDetachDisk struct {
	vmID string
}

func (s *stepDetachDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*openapi.APIClient)
	ui := state.Get("ui").(packer.Ui)
	vmID := state.Get("vm_id").(string)
	chrootDiskID := state.Get("chroot_disk_id").(string)

	ui.Say("Detaching chroot disk...")
	_, _, err := client.VmApi.VmDeleteHddDiskId(ctx, vmID, chrootDiskID)
	if err != nil {
		err := fmt.Errorf("error detaching disk: %s", formatOpenAPIError(err))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepDetachDisk) Cleanup(state multistep.StateBag) {}
