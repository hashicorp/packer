package vsphere_tpl
import (
	"context"
	"net/url"

	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
)

type StepVSphereClient struct {
	Url        *url.URL
	Datacenter string
	VMName     string
}

func (s *StepVSphereClient) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ctx := context.Background()
	cli, err := govmomi.NewClient(ctx, s.Url, true)

	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

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
	state.Put("context", ctx)
	return multistep.ActionContinue
}

func (s *StepVSphereClient) Cleanup(multistep.StateBag) {}