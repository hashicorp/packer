package chroot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"

	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepVerifySourceDisk struct {
	SourceDiskResourceID string
	Location             string
}

func (s StepVerifySourceDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	azcli := state.Get("azureclient").(client.AzureClientSet)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Checking source disk location")
	resource, err := azure.ParseResourceID(s.SourceDiskResourceID)
	if err != nil {
		log.Printf("StepVerifySourceDisk.Run: error: %+v", err)
		err := fmt.Errorf("Could not parse resource id %q: %s", s.SourceDiskResourceID, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if !strings.EqualFold(resource.SubscriptionID, azcli.SubscriptionID()) {
		err := fmt.Errorf("Source disk resource %q is in a different subscription than this VM (%q). "+
			"Packer does not know how to handle that.",
			s.SourceDiskResourceID, azcli.SubscriptionID())
		log.Printf("StepVerifySourceDisk.Run: error: %+v", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if !(strings.EqualFold(resource.Provider, "Microsoft.Compute") && strings.EqualFold(resource.ResourceType, "disks")) {
		err := fmt.Errorf("Resource ID %q is not a managed disk resource", s.SourceDiskResourceID)
		log.Printf("StepVerifySourceDisk.Run: error: %+v", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	disk, err := azcli.DisksClient().Get(ctx,
		resource.ResourceGroup, resource.ResourceName)
	if err != nil {
		err := fmt.Errorf("Unable to retrieve disk (%q): %s", s.SourceDiskResourceID, err)
		log.Printf("StepVerifySourceDisk.Run: error: %+v", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	location := to.String(disk.Location)
	if !strings.EqualFold(location, s.Location) {
		err := fmt.Errorf("Source disk resource %q is in a different location (%q) than this VM (%q). "+
			"Packer does not know how to handle that.",
			s.SourceDiskResourceID,
			location,
			s.Location)
		log.Printf("StepVerifySourceDisk.Run: error: %+v", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s StepVerifySourceDisk) Cleanup(state multistep.StateBag) {}
