package chroot

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"log"
)

var _ multistep.Step = &StepCreateImage{}

type StepCreateImage struct {
	ImageResourceID          string
	ImageOSState             string
	OSDiskStorageAccountType string
	OSDiskCacheType          string

	imageResource azure.Resource
}

func (s *StepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	azcli := state.Get("azureclient").(client.AzureClientSet)
	ui := state.Get("ui").(packer.Ui)
	diskResourceID := state.Get("os_disk_resource_id").(string)

	ui.Say(fmt.Sprintf("Creating image %s\n   using %s for os disk.",
		s.ImageResourceID,
		diskResourceID))

	var err error
	s.imageResource, err = azure.ParseResourceID(s.ImageResourceID)

	if err != nil {
		log.Printf("StepCreateImage.Run: error: %+v", err)
		err := fmt.Errorf(
			"error parsing image resource id '%s': %v", s.ImageResourceID, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	image := compute.Image{
		ImageProperties: &compute.ImageProperties{
			StorageProfile: &compute.ImageStorageProfile{
				OsDisk: &compute.ImageOSDisk{
					OsType:  "Linux",
					OsState: compute.OperatingSystemStateTypes(s.ImageOSState),
					ManagedDisk: &compute.SubResource{
						ID: &diskResourceID,
					},
					Caching:            compute.CachingTypes(s.OSDiskCacheType),
					StorageAccountType: compute.StorageAccountTypes(s.OSDiskStorageAccountType),
				},
				//	DataDisks:     nil,
				//	ZoneResilient: nil,
			},
		},
		//		Tags:            nil,
	}
	f, err := azcli.ImagesClient().CreateOrUpdate(
		ctx,
		s.imageResource.ResourceGroup,
		s.imageResource.ResourceName,
		image)
	if err == nil {
		err = f.WaitForCompletionRef(ctx, azcli.PollClient())
	}
	if err != nil {
		log.Printf("StepCreateImage.Run: error: %+v", err)
		err := fmt.Errorf(
			"error creating image '%s': %v", s.ImageResourceID, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepCreateImage) Cleanup(state multistep.StateBag) {
	azcli := state.Get("azureclient").(client.AzureClientSet)
	ui := state.Get("ui").(packer.Ui)

	ctx := context.Background()
	f, err := azcli.ImagesClient().Delete(
		ctx,
		s.imageResource.ResourceGroup,
		s.imageResource.ResourceName)
	if err == nil {
		err = f.WaitForCompletionRef(ctx, azcli.PollClient())
	}
	if err != nil {
		log.Printf("StepCreateImage.Cleanup: error: %+v", err)
		ui.Error(fmt.Sprintf(
			"error deleting image '%s': %v", s.ImageResourceID, err))
	}
}
