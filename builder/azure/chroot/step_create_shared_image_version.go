package chroot

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepCreateSharedImageVersion struct {
	Destination     SharedImageGalleryDestination
	OSDiskCacheType string
	Location        string
}

func (s *StepCreateSharedImageVersion) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	azcli := state.Get("azureclient").(client.AzureClientSet)
	ui := state.Get("ui").(packer.Ui)
	osDiskSnapshotResourceID := state.Get(stateBagKey_OSDiskSnapshotResourceID).(string)

	ui.Say(fmt.Sprintf("Creating image version %s\n   using %s for os disk.",
		s.Destination.ResourceID(azcli.SubscriptionID()),
		osDiskSnapshotResourceID))

	var targetRegions []compute.TargetRegion
	// transform target regions to API objects
	for _, tr := range s.Destination.TargetRegions {
		apiObject := compute.TargetRegion{
			Name:                 to.StringPtr(tr.Name),
			RegionalReplicaCount: to.Int32Ptr(tr.ReplicaCount),
			StorageAccountType:   compute.StorageAccountType(tr.StorageAccountType),
		}
		targetRegions = append(targetRegions, apiObject)
	}

	imageVersion := compute.GalleryImageVersion{
		Location: to.StringPtr(s.Location),
		GalleryImageVersionProperties: &compute.GalleryImageVersionProperties{
			StorageProfile: &compute.GalleryImageVersionStorageProfile{
				OsDiskImage: &compute.GalleryOSDiskImage{
					Source:      &compute.GalleryArtifactVersionSource{ID: &osDiskSnapshotResourceID},
					HostCaching: compute.HostCaching(s.OSDiskCacheType),
				},
			},
			PublishingProfile: &compute.GalleryImageVersionPublishingProfile{
				TargetRegions:     &targetRegions,
				ExcludeFromLatest: to.BoolPtr(s.Destination.ExcludeFromLatest),
			},
		},
	}

	f, err := azcli.GalleryImageVersionsClient().CreateOrUpdate(
		ctx,
		s.Destination.ResourceGroup,
		s.Destination.GalleryName,
		s.Destination.ImageName,
		s.Destination.ImageVersion,
		imageVersion)
	if err == nil {
		log.Println("Shared image version creation in process...")
		pollClient := azcli.PollClient()
		pollClient.PollingDelay = 10 * time.Second
		ctx, cancel := context.WithTimeout(ctx, time.Hour*12)
		defer cancel()
		err = f.WaitForCompletionRef(ctx, pollClient)
	}
	if err != nil {
		log.Printf("StepCreateSharedImageVersion.Run: error: %+v", err)
		err := fmt.Errorf(
			"error creating shared image version '%s': %v", s.Destination.ResourceID(azcli.SubscriptionID()), err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	log.Printf("Image creation complete: %s", f.Status())

	return multistep.ActionContinue
}

func (*StepCreateSharedImageVersion) Cleanup(multistep.StateBag) {}
