package arm

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/azure/common/constants"
)

type StepPublishToSharedImageGallery struct {
	client  *AzureClient
	publish func(ctx context.Context, mdiID string, sharedImageGallery SharedImageGalleryDestination, miSGImageVersionEndOfLifeDate string, miSGImageVersionExcludeFromLatest bool, miSigReplicaCount int32, location string, tags map[string]*string) (string, error)
	say     func(message string)
	error   func(e error)
	toSIG   func() bool
}

func NewStepPublishToSharedImageGallery(client *AzureClient, ui packersdk.Ui, config *Config) *StepPublishToSharedImageGallery {
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

func getSigDestination(state multistep.StateBag) SharedImageGalleryDestination {
	subscription := state.Get(constants.ArmManagedImageSubscription).(string)
	resourceGroup := state.Get(constants.ArmManagedImageSigPublishResourceGroup).(string)
	galleryName := state.Get(constants.ArmManagedImageSharedGalleryName).(string)
	imageName := state.Get(constants.ArmManagedImageSharedGalleryImageName).(string)
	imageVersion := state.Get(constants.ArmManagedImageSharedGalleryImageVersion).(string)
	replicationRegions := state.Get(constants.ArmManagedImageSharedGalleryReplicationRegions).([]string)

	return SharedImageGalleryDestination{
		SigDestinationSubscription:       subscription,
		SigDestinationResourceGroup:      resourceGroup,
		SigDestinationGalleryName:        galleryName,
		SigDestinationImageName:          imageName,
		SigDestinationImageVersion:       imageVersion,
		SigDestinationReplicationRegions: replicationRegions,
	}
}

func (s *StepPublishToSharedImageGallery) publishToSig(ctx context.Context, mdiID string, sharedImageGallery SharedImageGalleryDestination, miSGImageVersionEndOfLifeDate string, miSGImageVersionExcludeFromLatest bool, miSigReplicaCount int32, location string, tags map[string]*string) (string, error) {
	replicationRegions := make([]compute.TargetRegion, len(sharedImageGallery.SigDestinationReplicationRegions))
	for i, v := range sharedImageGallery.SigDestinationReplicationRegions {
		regionName := v
		replicationRegions[i] = compute.TargetRegion{Name: &regionName}
	}

	var endOfLifeDate *date.Time
	if miSGImageVersionEndOfLifeDate != "" {
		parseDate, err := date.ParseTime("2006-01-02T15:04:05.99Z", miSGImageVersionEndOfLifeDate)
		if err != nil {
			s.say(fmt.Sprintf("Error parsing date from shared_gallery_image_version_end_of_life_date: %s", err))
			return "", err
		}
		endOfLifeDate = &date.Time{Time: parseDate}
	} else {
		endOfLifeDate = (*date.Time)(nil)
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
				TargetRegions:     &replicationRegions,
				EndOfLifeDate:     endOfLifeDate,
				ExcludeFromLatest: &miSGImageVersionExcludeFromLatest,
				ReplicaCount:      &miSigReplicaCount,
			},
		},
	}

	f, err := s.client.GalleryImageVersionsClient.CreateOrUpdate(ctx, sharedImageGallery.SigDestinationResourceGroup, sharedImageGallery.SigDestinationGalleryName, sharedImageGallery.SigDestinationImageName, sharedImageGallery.SigDestinationImageVersion, galleryImageVersion)

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

	location := stateBag.Get(constants.ArmLocation).(string)
	tags := stateBag.Get(constants.ArmTags).(map[string]*string)

	sharedImageGallery := getSigDestination(stateBag)
	targetManagedImageResourceGroupName := stateBag.Get(constants.ArmManagedImageResourceGroupName).(string)
	targetManagedImageName := stateBag.Get(constants.ArmManagedImageName).(string)

	managedImageSubscription := stateBag.Get(constants.ArmManagedImageSubscription).(string)
	mdiID := fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/images/%s", managedImageSubscription, targetManagedImageResourceGroupName, targetManagedImageName)

	miSGImageVersionEndOfLifeDate, _ := stateBag.Get(constants.ArmManagedImageSharedGalleryImageVersionEndOfLifeDate).(string)
	miSGImageVersionExcludeFromLatest, _ := stateBag.Get(constants.ArmManagedImageSharedGalleryImageVersionExcludeFromLatest).(bool)
	miSigReplicaCount, _ := stateBag.Get(constants.ArmManagedImageSharedGalleryImageVersionReplicaCount).(int32)
	// Replica count must be between 1 and 10 inclusive.
	if miSigReplicaCount <= 0 {
		miSigReplicaCount = constants.SharedImageGalleryImageVersionDefaultMinReplicaCount
	} else if miSigReplicaCount > 10 {
		miSigReplicaCount = constants.SharedImageGalleryImageVersionDefaultMaxReplicaCount
	}

	s.say(fmt.Sprintf(" -> MDI ID used for SIG publish           : '%s'", mdiID))
	s.say(fmt.Sprintf(" -> SIG publish resource group            : '%s'", sharedImageGallery.SigDestinationResourceGroup))
	s.say(fmt.Sprintf(" -> SIG gallery name                      : '%s'", sharedImageGallery.SigDestinationGalleryName))
	s.say(fmt.Sprintf(" -> SIG image name                        : '%s'", sharedImageGallery.SigDestinationImageName))
	s.say(fmt.Sprintf(" -> SIG image version                     : '%s'", sharedImageGallery.SigDestinationImageVersion))
	s.say(fmt.Sprintf(" -> SIG replication regions               : '%v'", sharedImageGallery.SigDestinationReplicationRegions))
	s.say(fmt.Sprintf(" -> SIG image version endoflife date      : '%s'", miSGImageVersionEndOfLifeDate))
	s.say(fmt.Sprintf(" -> SIG image version exclude from latest : '%t'", miSGImageVersionExcludeFromLatest))
	s.say(fmt.Sprintf(" -> SIG replica count [1, 10]             : '%d'", miSigReplicaCount))

	createdGalleryImageVersionID, err := s.publish(ctx, mdiID, sharedImageGallery, miSGImageVersionEndOfLifeDate, miSGImageVersionExcludeFromLatest, miSigReplicaCount, location, tags)

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
