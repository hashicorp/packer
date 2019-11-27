package arm

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"
	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepPublishToSharedImageGallery struct {
	client  *AzureClient
	publish func(ctx context.Context, mdiID, miSigPubRg, miSIGalleryName, miSGImageName, miSGImageVersion string, miSigReplicationRegions []string, location string, tags map[string]*string) (string, error)
	say     func(message string)
	error   func(e error)
	toSIG   func() bool
}

func NewStepPublishToSharedImageGallery(client *AzureClient, ui packer.Ui, config *Config) *StepPublishToSharedImageGallery {
	var step = &StepPublishToSharedImageGallery{
		client: client,
		say: func(message string) {
			ui.Say(message)
		},
		error: func(e error) {
			ui.Error(e.Error())
		},
		toSIG: func() bool {
			return config.isManagedImage() && config.SharedGalleryDestination.SigDestinationGalleryName != ""
		},
	}

	step.publish = step.publishToSig
	return step
}

func (s *StepPublishToSharedImageGallery) publishToSig(ctx context.Context, mdiID string, miSigPubRg string, miSIGalleryName string, miSGImageName string, miSGImageVersion string, miSigReplicationRegions []string, location string, tags map[string]*string) (string, error) {

	replicationRegions := make([]compute.TargetRegion, len(miSigReplicationRegions))
	for i, v := range miSigReplicationRegions {
		regionName := v
		replicationRegions[i] = compute.TargetRegion{Name: &regionName}
	}

	galleryImageVersion := compute.GalleryImageVersion{
		Location: &location,
		Tags:     tags,
		GalleryImageVersionProperties: &compute.GalleryImageVersionProperties{
			PublishingProfile: &compute.GalleryImageVersionPublishingProfile{
				Source: &compute.GalleryArtifactSource{
					ManagedImage: &compute.ManagedArtifact{
						ID: &mdiID,
					},
				},
				TargetRegions: &replicationRegions,
			},
		},
	}

	f, err := s.client.GalleryImageVersionsClient.CreateOrUpdate(ctx, miSigPubRg, miSIGalleryName, miSGImageName, miSGImageVersion, galleryImageVersion)

	if err != nil {
		s.say(s.client.LastError.Error())
		return "", err
	}

	err = f.WaitForCompletionRef(ctx, s.client.GalleryImageVersionsClient.Client)

	if err != nil {
		s.say(s.client.LastError.Error())
		return "", err
	}

	createdSGImageVersion, err := f.Result(s.client.GalleryImageVersionsClient)

	if err != nil {
		s.say(s.client.LastError.Error())
		return "", err
	}

	s.say(fmt.Sprintf(" -> Shared Gallery Image Version ID : '%s'", *(createdSGImageVersion.ID)))
	return *(createdSGImageVersion.ID), nil
}

func (s *StepPublishToSharedImageGallery) Run(ctx context.Context, stateBag multistep.StateBag) multistep.StepAction {
	if !s.toSIG() {
		return multistep.ActionContinue
	}

	s.say("Publishing to Shared Image Gallery ...")

	var miSigPubRg = stateBag.Get(constants.ArmManagedImageSigPublishResourceGroup).(string)
	var miSIGalleryName = stateBag.Get(constants.ArmManagedImageSharedGalleryName).(string)
	var miSGImageName = stateBag.Get(constants.ArmManagedImageSharedGalleryImageName).(string)
	var miSGImageVersion = stateBag.Get(constants.ArmManagedImageSharedGalleryImageVersion).(string)
	var location = stateBag.Get(constants.ArmLocation).(string)
	var tags = stateBag.Get(constants.ArmTags).(map[string]*string)
	var miSigReplicationRegions = stateBag.Get(constants.ArmManagedImageSharedGalleryReplicationRegions).([]string)
	var targetManagedImageResourceGroupName = stateBag.Get(constants.ArmManagedImageResourceGroupName).(string)
	var targetManagedImageName = stateBag.Get(constants.ArmManagedImageName).(string)
	var managedImageSubscription = stateBag.Get(constants.ArmManagedImageSubscription).(string)
	var mdiID = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/images/%s", managedImageSubscription, targetManagedImageResourceGroupName, targetManagedImageName)

	s.say(fmt.Sprintf(" -> MDI ID used for SIG publish     : '%s'", mdiID))
	s.say(fmt.Sprintf(" -> SIG publish resource group     : '%s'", miSigPubRg))
	s.say(fmt.Sprintf(" -> SIG gallery name     : '%s'", miSIGalleryName))
	s.say(fmt.Sprintf(" -> SIG image name     : '%s'", miSGImageName))
	s.say(fmt.Sprintf(" -> SIG image version     : '%s'", miSGImageVersion))
	s.say(fmt.Sprintf(" -> SIG replication regions    : '%v'", miSigReplicationRegions))
	createdGalleryImageVersionID, err := s.publish(ctx, mdiID, miSigPubRg, miSIGalleryName, miSGImageName, miSGImageVersion, miSigReplicationRegions, location, tags)

	if err != nil {
		stateBag.Put(constants.Error, err)
		s.error(err)

		return multistep.ActionHalt
	}

	stateBag.Put(constants.ArmManagedImageSharedGalleryId, createdGalleryImageVersionID)
	return multistep.ActionContinue
}

func (*StepPublishToSharedImageGallery) Cleanup(multistep.StateBag) {
}
