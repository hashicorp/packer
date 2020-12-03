package hyperone

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	openapi "github.com/hyperonecom/h1-client-go"
)

type stepCreateVMFromDisk struct {
	vmID string
}

func (s *stepCreateVMFromDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*openapi.APIClient)
	ui := state.Get("ui").(packersdk.Ui)
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

	ui := state.Get("ui").(packersdk.Ui)

	ui.Say(fmt.Sprintf("Deleting VM %s (from chroot disk)...", s.vmID))
	err := deleteVMWithDisks(s.vmID, state)
	if err != nil {
		ui.Error(err.Error())
	}
}
