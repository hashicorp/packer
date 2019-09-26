package chroot

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"

	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepVerifySourceDisk struct {
	SubscriptionID       string
	SourceDiskResourceID string
	Location             string
}

func (s StepVerifySourceDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	azcli := state.Get("azureclient").(client.AzureClientSet)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Checking source disk location")
	resource, err := azure.ParseResourceID(s.SourceDiskResourceID)
	if err != nil {
		ui.Error(fmt.Sprintf("Could not parse resource id %q: %s", s.SourceDiskResourceID, err))
		return multistep.ActionHalt
	}

	if !strings.EqualFold(resource.SubscriptionID, s.SubscriptionID) {
		ui.Error(fmt.Sprintf("Source disk resource %q is in a different subscription than this VM (%q). "+
			"Packer does not know how to handle that.",
			s.SourceDiskResourceID, s.SubscriptionID))
		return multistep.ActionHalt
	}

	if !(strings.EqualFold(resource.Provider, "Microsoft.Compute") && strings.EqualFold(resource.ResourceType, "disks")) {
		ui.Error(fmt.Sprintf("Resource ID %q is not a managed disk resource", s.SourceDiskResourceID))
		return multistep.ActionHalt
	}

	disk, err := azcli.DisksClient().Get(ctx,
		resource.ResourceGroup, resource.ResourceName)
	if err != nil {
		ui.Error(fmt.Sprintf("Unable to retrieve disk (%q): %s", s.SourceDiskResourceID, err))
		return multistep.ActionHalt
	}

	location := to.String(disk.Location)
	if !strings.EqualFold(location, s.Location) {
		ui.Error(fmt.Sprintf("Source disk resource %q is in a different location (%q) than this VM (%q). "+
			"Packer does not know how to handle that.",
			s.SourceDiskResourceID,
			location,
			s.Location))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s StepVerifySourceDisk) Cleanup(state multistep.StateBag) {}
