package chroot

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

var _ multistep.Step = &StepCreateNewDisk{}

type StepCreateNewDisk struct {
	SubscriptionID, ResourceGroup, DiskName string
	DiskSizeGB                              int32  // optional, ignored if 0
	DiskStorageAccountType                  string // from compute.DiskStorageAccountTypes
	HyperVGeneration                        string

	Location      string
	PlatformImage *client.PlatformImage

	SourceDiskResourceID string

	SkipCleanup bool
}

func (s StepCreateNewDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	azcli := state.Get("azureclient").(client.AzureClientSet)
	ui := state.Get("ui").(packer.Ui)

	diskResourceID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/disks/%s",
		s.SubscriptionID,
		s.ResourceGroup,
		s.DiskName)
	state.Put("os_disk_resource_id", diskResourceID)
	ui.Say(fmt.Sprintf("Creating disk '%s'", diskResourceID))

	disk := compute.Disk{
		Location: to.StringPtr(s.Location),
		Sku: &compute.DiskSku{
			Name: compute.DiskStorageAccountTypes(s.DiskStorageAccountType),
		},
		//Zones: nil,
		DiskProperties: &compute.DiskProperties{
			OsType:           "Linux",
			HyperVGeneration: compute.HyperVGeneration(s.HyperVGeneration),
			CreationData:     &compute.CreationData{},
		},
		//Tags: map[string]*string{
	}

	if s.DiskSizeGB > 0 {
		disk.DiskProperties.DiskSizeGB = to.Int32Ptr(s.DiskSizeGB)
	}

	if s.SourceDiskResourceID != "" {
		disk.CreationData.CreateOption = compute.Copy
		disk.CreationData.SourceResourceID = to.StringPtr(s.SourceDiskResourceID)
	} else if s.PlatformImage == nil {
		disk.CreationData.CreateOption = compute.Empty
	} else {
		disk.CreationData.CreateOption = compute.FromImage
		disk.CreationData.ImageReference = &compute.ImageDiskReference{
			ID: to.StringPtr(fmt.Sprintf(
				"/subscriptions/%s/providers/Microsoft.Compute/locations/%s/publishers/%s/artifacttypes/vmimage/offers/%s/skus/%s/versions/%s",
				s.SubscriptionID, s.Location, s.PlatformImage.Publisher, s.PlatformImage.Offer, s.PlatformImage.Sku, s.PlatformImage.Version)),
		}
	}

	f, err := azcli.DisksClient().CreateOrUpdate(ctx, s.ResourceGroup, s.DiskName, disk)
	if err == nil {
		err = f.WaitForCompletionRef(ctx, azcli.PollClient())
	}
	if err != nil {
		log.Printf("StepCreateNewDisk.Run: error: %+v", err)
		err := fmt.Errorf(
			"error creating new disk '%s': %v", diskResourceID, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s StepCreateNewDisk) Cleanup(state multistep.StateBag) {
	if !s.SkipCleanup {
		azcli := state.Get("azureclient").(client.AzureClientSet)
		ui := state.Get("ui").(packer.Ui)
		diskResourceID := state.Get("os_disk_resource_id").(string)

		ui.Say(fmt.Sprintf("Waiting for disk %q detach to complete", diskResourceID))
		err := NewDiskAttacher(azcli).WaitForDetach(context.Background(), diskResourceID)
		if err != nil {
			ui.Error(fmt.Sprintf("error detaching disk %q: %s", diskResourceID, err))
		}

		ui.Say(fmt.Sprintf("Deleting disk %q", diskResourceID))

		f, err := azcli.DisksClient().Delete(context.TODO(), s.ResourceGroup, s.DiskName)
		if err == nil {
			err = f.WaitForCompletionRef(context.TODO(), azcli.PollClient())
		}
		if err != nil {
			log.Printf("StepCreateNewDisk.Cleanup: error: %+v", err)
			ui.Error(fmt.Sprintf("error deleting new disk '%s': %v.", diskResourceID, err))
		}
	}
}
