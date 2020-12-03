package chroot

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

var _ multistep.Step = &StepCreateImage{}

type StepCreateImage struct {
	ImageResourceID            string
	ImageOSState               string
	OSDiskStorageAccountType   string
	OSDiskCacheType            string
	DataDiskStorageAccountType string
	DataDiskCacheType          string
	Location                   string
}

func (s *StepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	azcli := state.Get("azureclient").(client.AzureClientSet)
	ui := state.Get("ui").(packersdk.Ui)
	diskset := state.Get(stateBagKey_Diskset).(Diskset)
	diskResourceID := diskset.OS().String()

	ui.Say(fmt.Sprintf("Creating image %s\n   using %s for os disk.",
		s.ImageResourceID,
		diskResourceID))

	imageResource, err := azure.ParseResourceID(s.ImageResourceID)

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

	var datadisks []compute.ImageDataDisk
	for lun, resource := range diskset {
		if lun != -1 {
			ui.Say(fmt.Sprintf("   using %q for data disk (lun %d).", resource, lun))

			datadisks = append(datadisks, compute.ImageDataDisk{
				Lun:                to.Int32Ptr(lun),
				ManagedDisk:        &compute.SubResource{ID: to.StringPtr(resource.String())},
				StorageAccountType: compute.StorageAccountTypes(s.DataDiskStorageAccountType),
				Caching:            compute.CachingTypes(s.DataDiskCacheType),
			})
		}
	}
	if datadisks != nil {
		sort.Slice(datadisks, func(i, j int) bool {
			return *datadisks[i].Lun < *datadisks[j].Lun
		})
		image.ImageProperties.StorageProfile.DataDisks = &datadisks
	}

	f, err := azcli.ImagesClient().CreateOrUpdate(
		ctx,
		imageResource.ResourceGroup,
		imageResource.ResourceName,
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
