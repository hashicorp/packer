package chroot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
)

var _ multistep.Step = &StepCreateNewDiskset{}

type StepCreateNewDiskset struct {
	OSDiskID                   string // Disk ID
	OSDiskSizeGB               int32  // optional, ignored if 0
	OSDiskStorageAccountType   string // from compute.DiskStorageAccountTypes
	DataDiskStorageAccountType string // from compute.DiskStorageAccountTypes

	DataDiskIDPrefix string

	disks Diskset

	HyperVGeneration string // For OS disk

	// Copy another disk
	SourceOSDiskResourceID string

	// Extract from platform image
	SourcePlatformImage *client.PlatformImage
	// Extract from shared image
	SourceImageResourceID string
	// Location is needed for platform and shared images
	Location string

	SkipCleanup bool
}

func (s *StepCreateNewDiskset) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	azcli := state.Get("azureclient").(client.AzureClientSet)
	ui := state.Get("ui").(packersdk.Ui)

	s.disks = make(Diskset)

	errorMessage := func(format string, params ...interface{}) multistep.StepAction {
		err := fmt.Errorf("StepCreateNewDiskset.Run: error: "+format, params...)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// we always have an OS disk
	osDisk, err := client.ParseResourceID(s.OSDiskID)
	if err != nil {
		return errorMessage("error parsing resource id '%s': %v", s.OSDiskID, err)
	}
	if !strings.EqualFold(osDisk.Provider, "Microsoft.Compute") ||
		!strings.EqualFold(osDisk.ResourceType.String(), "disks") {
		return errorMessage("Resource %q is not of type Microsoft.Compute/disks", s.OSDiskID)
	}

	// transform step config to disk model
	disk := s.getOSDiskDefinition(azcli.SubscriptionID())

	// Initiate disk creation
	f, err := azcli.DisksClient().CreateOrUpdate(ctx, osDisk.ResourceGroup, osDisk.ResourceName.String(), disk)
	if err != nil {
		return errorMessage("Failed to initiate resource creation: %q", osDisk)
	}
	s.disks[-1] = osDisk                    // save the resoure we just create in our disk set
	state.Put(stateBagKey_Diskset, s.disks) // update the statebag
	ui.Say(fmt.Sprintf("Creating disk %q", s.OSDiskID))

	type Future struct {
		client.Resource
		compute.DisksCreateOrUpdateFuture
	}
	futures := []Future{{osDisk, f}}

	if s.SourceImageResourceID != "" {
		// retrieve image to see if there are any datadisks
		imageID, err := client.ParseResourceID(s.SourceImageResourceID)
		if err != nil {
			return errorMessage("could not parse source image id %q: %v", s.SourceImageResourceID, err)
		}
		if !strings.EqualFold(imageID.Provider+"/"+imageID.ResourceType.String(),
			"Microsoft.Compute/galleries/images/versions") {
			return errorMessage("source image id is not a shared image version %q, expected type 'Microsoft.Compute/galleries/images/versions'", imageID)
		}
		image, err := azcli.GalleryImageVersionsClient().Get(ctx,
			imageID.ResourceGroup,
			imageID.ResourceName[0], imageID.ResourceName[1], imageID.ResourceName[2], "")
		if err != nil {
			return errorMessage("error retrieving source image %q: %v", imageID, err)
		}
		if image.GalleryImageVersionProperties != nil &&
			image.GalleryImageVersionProperties.StorageProfile != nil &&
			image.GalleryImageVersionProperties.StorageProfile.DataDiskImages != nil {
			for i, ddi := range *image.GalleryImageVersionProperties.StorageProfile.DataDiskImages {
				if ddi.Lun == nil {
					return errorMessage("unexpected: lun is null for data disk # %d", i)
				}
				datadiskID, err := client.ParseResourceID(fmt.Sprintf("%s%d", s.DataDiskIDPrefix, *ddi.Lun))
				if err != nil {
					return errorMessage("unable to construct resource id for datadisk: %v", err)
				}

				disk := s.getDatadiskDefinitionFromImage(*ddi.Lun)
				// Initiate disk creation
				f, err := azcli.DisksClient().CreateOrUpdate(ctx, datadiskID.ResourceGroup, datadiskID.ResourceName.String(), disk)
				if err != nil {
					return errorMessage("Failed to initiate resource creation: %q", datadiskID)
				}
				s.disks[*ddi.Lun] = datadiskID          // save the resoure we just create in our disk set
				state.Put(stateBagKey_Diskset, s.disks) // update the statebag
				ui.Say(fmt.Sprintf("Creating disk %q", datadiskID))

				futures = append(futures, Future{datadiskID, f})
			}
		}
	}

	ui.Say("Waiting for disks to be created.")

	// Wait for completion
	for _, f := range futures {
		cli := azcli.PollClient() // quick polling for quick operations
		cli.PollingDelay = time.Second
		err = f.WaitForCompletionRef(ctx, cli)
		if err != nil {
			return errorMessage(
				"error creating new disk '%s': %v", f.Resource, err)
		}
		ui.Say(fmt.Sprintf("Disk %q created", f.Resource))
	}

	return multistep.ActionContinue
}

func (s StepCreateNewDiskset) getOSDiskDefinition(subscriptionID string) compute.Disk {
	disk := compute.Disk{
		Location: to.StringPtr(s.Location),
		DiskProperties: &compute.DiskProperties{
			OsType:       "Linux",
			CreationData: &compute.CreationData{},
		},
	}

	if s.OSDiskStorageAccountType != "" {
		disk.Sku = &compute.DiskSku{
			Name: compute.DiskStorageAccountTypes(s.OSDiskStorageAccountType),
		}
	}

	if s.HyperVGeneration != "" {
		disk.DiskProperties.HyperVGeneration = compute.HyperVGeneration(s.HyperVGeneration)
	}

	if s.OSDiskSizeGB > 0 {
		disk.DiskProperties.DiskSizeGB = to.Int32Ptr(s.OSDiskSizeGB)
	}

	switch {
	case s.SourcePlatformImage != nil:
		disk.CreationData.CreateOption = compute.FromImage
		disk.CreationData.ImageReference = &compute.ImageDiskReference{
			ID: to.StringPtr(fmt.Sprintf(
				"/subscriptions/%s/providers/Microsoft.Compute/locations/%s/publishers/%s/artifacttypes/vmimage/offers/%s/skus/%s/versions/%s",
				subscriptionID, s.Location,
				s.SourcePlatformImage.Publisher, s.SourcePlatformImage.Offer, s.SourcePlatformImage.Sku, s.SourcePlatformImage.Version)),
		}
	case s.SourceOSDiskResourceID != "":
		disk.CreationData.CreateOption = compute.Copy
		disk.CreationData.SourceResourceID = to.StringPtr(s.SourceOSDiskResourceID)
	case s.SourceImageResourceID != "":
		disk.CreationData.CreateOption = compute.FromImage
		disk.CreationData.GalleryImageReference = &compute.ImageDiskReference{
			ID: to.StringPtr(s.SourceImageResourceID),
		}
	default:
		disk.CreationData.CreateOption = compute.Empty
	}
	return disk
}

func (s StepCreateNewDiskset) getDatadiskDefinitionFromImage(lun int32) compute.Disk {
	disk := compute.Disk{
		Location: to.StringPtr(s.Location),
		DiskProperties: &compute.DiskProperties{
			CreationData: &compute.CreationData{},
		},
	}

	disk.CreationData.CreateOption = compute.FromImage
	disk.CreationData.GalleryImageReference = &compute.ImageDiskReference{
		ID:  to.StringPtr(s.SourceImageResourceID),
		Lun: to.Int32Ptr(lun),
	}

	if s.DataDiskStorageAccountType != "" {
		disk.Sku = &compute.DiskSku{
			Name: compute.DiskStorageAccountTypes(s.DataDiskStorageAccountType),
		}
	}
	return disk
}

func (s *StepCreateNewDiskset) Cleanup(state multistep.StateBag) {
	if !s.SkipCleanup {
		azcli := state.Get("azureclient").(client.AzureClientSet)
		ui := state.Get("ui").(packersdk.Ui)

		for _, d := range s.disks {

			ui.Say(fmt.Sprintf("Waiting for disk %q detach to complete", d))
			err := NewDiskAttacher(azcli).WaitForDetach(context.Background(), d.String())
			if err != nil {
				ui.Error(fmt.Sprintf("error detaching disk %q: %s", d, err))
			}

			ui.Say(fmt.Sprintf("Deleting disk %q", d))

			f, err := azcli.DisksClient().Delete(context.TODO(), d.ResourceGroup, d.ResourceName.String())
			if err == nil {
				err = f.WaitForCompletionRef(context.TODO(), azcli.PollClient())
			}
			if err != nil {
				log.Printf("StepCreateNewDiskset.Cleanup: error: %+v", err)
				ui.Error(fmt.Sprintf("error deleting disk '%s': %v.", d, err))
			}
		}
	}
}
