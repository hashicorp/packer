package vsphere_template

import (
	"context"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
)

type StepChooseDatacenter struct {
	Datacenter string
}

func (s *StepChooseDatacenter) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	cli := state.Get("client").(*govmomi.Client)
	ctx := state.Get("context").(context.Context)
	finder := find.NewFinder(cli.Client, false)

	datacenter, err := finder.DatacenterOrDefault(ctx, s.Datacenter)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())

		return multistep.ActionHalt
	}

	finder.SetDatacenter(datacenter)
	state.Put("datacenter", datacenter.Name())
	state.Put("finder", finder)
	return multistep.ActionContinue
}

func (s *StepChooseDatacenter) Cleanup(multistep.StateBag) {}
