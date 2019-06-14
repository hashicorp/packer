package arm

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"
	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepPublishToSharedImageGallery struct {
	client              *AzureClient
	publish             func(ctx context.Context, miSigPubSubscription, miSigPubRg, miSIGalleryName, miSGImageName, miSGImageVersion string, miSigReplicationRegions []string, location string, tags map[string]*string) error
	say                 func(message string)
	error               func(e error)
	toSIG               func() bool
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

func (s *StepPublishToSharedImageGallery) publishToSig(ctx context.Context, miSigPubSubscription string, miSigPubRg string, miSIGalleryName string, miSGImageName string, miSGImageVersion string, miSigReplicationRegions []string, location string, tags map[string]*string) error {

	var mdiID string

	replicationRegions := make([]compute.TargetRegion, len(miSigReplicationRegions))
	for i, v := range miSigReplicationRegions {
		regionName := v
		replicationRegions[i] = compute.TargetRegion{Name: &regionName}
	}

	galleryImageVersion := compute.GalleryImageVersion{
		Location: &location,
		Tags: tags,
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

	timeStamp := time.Now().Unix() // returns 64 bit value
	versionName := "0."
	// get the higher 16 bits
	minorVersion := timeStamp >> 16
	// get the lower 16 bits
	patch := timeStamp & 0xffff
	versionName += strconv.FormatInt(minorVersion, 10) + "." + strconv.FormatInt(patch, 10)

	f, err := s.client.GalleryImageVersionsClient.CreateOrUpdate(ctx, miSigPubRg, miSIGalleryName, miSGImageName, versionName, galleryImageVersion)

	if err != nil {
		s.say(s.client.LastError.Error())
		return err
	}

	err = f.WaitForCompletionRef(ctx, s.client.GalleryImageVersionsClient.Client)

	if err != nil {
		s.say(s.client.LastError.Error())
		return err
	}

	createdSGImageVersion, err := f.Result(s.client.GalleryImageVersionsClient)

	if err != nil {
		s.say(s.client.LastError.Error())
		return err
	}

	// TODO: compare and contrast to see if Amrita needs to add this to artifact id
	s.say(fmt.Sprintf(" -> Shared Gallery Image Version ID : '%s'", *(createdSGImageVersion.ID)))
	return nil
}

func (s *StepPublishToSharedImageGallery) Run(ctx context.Context, stateBag multistep.StateBag) multistep.StepAction {
	s.say("Publishing to Shared Image Gallery ...")

	var miSigPubSubscription = stateBag.Get(constants.ArmManagedImageSigPublishSubscription).(string)
	var miSigPubRg = stateBag.Get(constants.ArmManagedImageSigPublishResourceGroup).(string)
	var miSIGalleryName = stateBag.Get(constants.ArmManagedImageSharedGalleryName).(string)
	var miSGImageName = stateBag.Get(constants.ArmManagedImageSharedGalleryImageName).(string)
	var miSGImageVersion = stateBag.Get(constants.ArmManagedImageSharedGalleryImageVersion).(string)
	var location = stateBag.Get(constants.ArmLocation).(string)
	var tags = stateBag.Get(constants.ArmTags).(map[string]*string)
	var miSigReplicationRegions = stateBag.Get(constants.ArmManagedImageSharedGalleryReplicationRegions).([]string)

	s.say(fmt.Sprintf(" -> SIG publish subscription     : '%s'", miSigPubSubscription))
	s.say(fmt.Sprintf(" -> SIG publish resource group     : '%s'", miSigPubRg))
	s.say(fmt.Sprintf(" -> SIG gallery name     : '%s'", miSIGalleryName))
	s.say(fmt.Sprintf(" -> SIG image name     : '%s'", miSGImageName))
	s.say(fmt.Sprintf(" -> SIG image version     : '%s'", miSGImageVersion))
	s.say(fmt.Sprintf(" -> SIG publish location     : '%s'", location))
	err := s.publish(ctx, miSigPubSubscription, miSigPubRg, miSIGalleryName, miSGImageName, miSGImageVersion, miSigReplicationRegions,location, tags)

	if err != nil {
		stateBag.Put(constants.Error, err)
		s.error(err)

		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (*StepPublishToSharedImageGallery) Cleanup(multistep.StateBag) {
}
