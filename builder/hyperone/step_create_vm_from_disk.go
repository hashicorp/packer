package hyperone

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hyperonecom/h1-client-go"
)

type stepCreateVMFromDisk struct {
	vmID string
}

func (s *stepCreateVMFromDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*openapi.APIClient)
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)
	sshKey := state.Get("ssh_public_key").(string)
	chrootDiskID := state.Get("chroot_disk_id").(string)

	ui.Say("Creating VM from disk...")

	options := openapi.VmCreate{
		Name:    config.VmName,
		Service: config.VmType,
		Disk: []openapi.VmCreateDisk{
			{
				Id: chrootDiskID,
			},
		},
		SshKeys: []string{sshKey},
		Boot:    false,
	}

	vm, _, err := client.VmApi.VmCreate(ctx, options)
	if err != nil {
		err := fmt.Errorf("error creating VM from disk: %s", formatOpenAPIError(err))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.vmID = vm.Id
	state.Put("vm_id", vm.Id)

	return multistep.ActionContinue
}

func (s *stepCreateVMFromDisk) Cleanup(state multistep.StateBag) {
	if s.vmID == "" {
		return
	}

	client := state.Get("client").(*openapi.APIClient)
	ui := state.Get("ui").(packer.Ui)
	chrootDiskID := state.Get("chroot_disk_id").(string)

	ui.Say("Deleting VM (from disk)...")

	deleteOptions := openapi.VmDelete{
		RemoveDisks: []string{chrootDiskID},
	}

	_, err := client.VmApi.VmDelete(context.TODO(), s.vmID, deleteOptions)
	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting server '%s' - please delete it manually: %s", s.vmID, formatOpenAPIError(err)))
	}
}
