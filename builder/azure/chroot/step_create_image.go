package chroot

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

var _ multistep.Step = &StepCreateImage{}

type StepCreateImage struct {
	ImageResourceID          string
	ImageOSState             string
	OSDiskStorageAccountType string
	OSDiskCacheType          string
	Location                 string

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
		Location: to.StringPtr(s.Location),
		ImageProperties: &compute.ImageProperties{
			StorageProfile: &compute.ImageStorageProfile{
				OsDisk: &compute.ImageOSDisk{
					OsState: compute.OperatingSystemStateTypes(s.ImageOSState),
					OsType:  compute.Linux,
					ManagedDisk: &compute.SubResource{
						ID: &diskResourceID,
					},
					StorageAccountType: compute.StorageAccountTypes(s.OSDiskStorageAccountType),
					Caching:            compute.CachingTypes(s.OSDiskCacheType),
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
		log.Println("Image creation in process...")
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
	log.Printf("Image creation complete: %s", f.Status())

	return multistep.ActionContinue
}

func (*StepCreateImage) Cleanup(bag multistep.StateBag) {} // this is the final artifact, don't delete
