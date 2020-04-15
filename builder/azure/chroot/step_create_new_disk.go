package chroot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

var _ multistep.Step = &StepCreateNewDisk{}

type StepCreateNewDisk struct {
	ResourceID                              string // Disk ID
	subscriptionID, resourceGroup, diskName string // split out resource id

	DiskSizeGB             int32  // optional, ignored if 0
	DiskStorageAccountType string // from compute.DiskStorageAccountTypes
	HyperVGeneration       string

	Location      string
	PlatformImage *client.PlatformImage

	SourceDiskResourceID string

	SourceImageResourceID string

	SkipCleanup bool
}

func parseDiskResourceID(resourceID string) (subscriptionID, resourceGroup, diskName string, err error) {
	r, err := azure.ParseResourceID(resourceID)
	if err != nil {
		return "", "", "", err
	}

	if !strings.EqualFold(r.Provider, "Microsoft.Compute") ||
		!strings.EqualFold(r.ResourceType, "disks") {
		return "", "", "", fmt.Errorf("Resource %q is not of type Microsoft.Compute/disks", resourceID)
	}

	return r.SubscriptionID, r.ResourceGroup, r.ResourceName, nil
}

func (s *StepCreateNewDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	azcli := state.Get("azureclient").(client.AzureClientSet)
	ui := state.Get("ui").(packer.Ui)

	state.Put(stateBagKey_OSDiskResourceID, s.ResourceID)
	ui.Say(fmt.Sprintf("Creating disk '%s'", s.ResourceID))

	var err error
	s.subscriptionID, s.resourceGroup, s.diskName, err = parseDiskResourceID(s.ResourceID)
	if err != nil {
		log.Printf("StepCreateNewDisk.Run: error: %+v", err)
		err := fmt.Errorf(
			"error parsing resource id '%s': %v", s.ResourceID, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	disk := compute.Disk{
		Location: to.StringPtr(s.Location),
		DiskProperties: &compute.DiskProperties{
			OsType:       "Linux",
			CreationData: &compute.CreationData{},
		},
	}

	if s.DiskStorageAccountType != "" {
		disk.Sku = &compute.DiskSku{
			Name: compute.DiskStorageAccountTypes(s.DiskStorageAccountType),
		}
	}

	if s.HyperVGeneration != "" {
		disk.DiskProperties.HyperVGeneration = compute.HyperVGeneration(s.HyperVGeneration)
	}

	if s.DiskSizeGB > 0 {
		disk.DiskProperties.DiskSizeGB = to.Int32Ptr(s.DiskSizeGB)
	}

	switch {
	case s.PlatformImage != nil:
		disk.CreationData.CreateOption = compute.FromImage
		disk.CreationData.ImageReference = &compute.ImageDiskReference{
			ID: to.StringPtr(fmt.Sprintf(
				"/subscriptions/%s/providers/Microsoft.Compute/locations/%s/publishers/%s/artifacttypes/vmimage/offers/%s/skus/%s/versions/%s",
				s.subscriptionID, s.Location, s.PlatformImage.Publisher, s.PlatformImage.Offer, s.PlatformImage.Sku, s.PlatformImage.Version)),
		}
	case s.SourceDiskResourceID != "":
		disk.CreationData.CreateOption = compute.Copy
		disk.CreationData.SourceResourceID = to.StringPtr(s.SourceDiskResourceID)
	case s.SourceImageResourceID != "":
		disk.CreationData.CreateOption = compute.FromImage
		disk.CreationData.GalleryImageReference = &compute.ImageDiskReference{
			ID: to.StringPtr(s.SourceImageResourceID),
		}
	default:
		disk.CreationData.CreateOption = compute.Empty
	}

	f, err := azcli.DisksClient().CreateOrUpdate(ctx, s.resourceGroup, s.diskName, disk)
	if err == nil {
		cli := azcli.PollClient() // quick polling for quick operations
		cli.PollingDelay = time.Second
		err = f.WaitForCompletionRef(ctx, cli)
	}
	if err != nil {
		log.Printf("StepCreateNewDisk.Run: error: %+v", err)
		err := fmt.Errorf(
			"error creating new disk '%s': %v", s.ResourceID, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepCreateNewDisk) Cleanup(state multistep.StateBag) {
	if !s.SkipCleanup {
		azcli := state.Get("azureclient").(client.AzureClientSet)
		ui := state.Get("ui").(packer.Ui)

		ui.Say(fmt.Sprintf("Waiting for disk %q detach to complete", s.ResourceID))
		err := NewDiskAttacher(azcli).WaitForDetach(context.Background(), s.ResourceID)
		if err != nil {
			ui.Error(fmt.Sprintf("error detaching disk %q: %s", s.ResourceID, err))
		}

		ui.Say(fmt.Sprintf("Deleting disk %q", s.ResourceID))

		f, err := azcli.DisksClient().Delete(context.TODO(), s.resourceGroup, s.diskName)
		if err == nil {
			err = f.WaitForCompletionRef(context.TODO(), azcli.PollClient())
		}
		if err != nil {
			log.Printf("StepCreateNewDisk.Cleanup: error: %+v", err)
			ui.Error(fmt.Sprintf("error deleting disk '%s': %v.", s.ResourceID, err))
		}
	}
}
