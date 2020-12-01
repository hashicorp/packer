package vsphere_template

import (
	"context"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
)

type stepChooseDatacenter struct {
	Datacenter string
}

func (s *stepChooseDatacenter) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	cli := state.Get("client").(*govmomi.Client)
	finder := find.NewFinder(cli.Client, false)

	ui.Message("Choosing datacenter...")

	dc, err := finder.DatacenterOrDefault(context.Background(), s.Datacenter)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("dcPath", dc.InventoryPath)

	return multistep.ActionContinue
}

func (s *stepChooseDatacenter) Cleanup(multistep.StateBag) {}
