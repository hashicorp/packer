package chroot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

var _ multistep.Step = &StepVerifySharedImageDestination{}

// StepVerifySharedImageDestination verifies that the shared image location matches the Location field in the step.
// Also verifies that the OS Type is Linux.
type StepVerifySharedImageDestination struct {
	Image    SharedImageGalleryDestination
	Location string
}

// Run retrieves the image metadata from Azure and compares the location to Location. Verifies the OS Type.
func (s *StepVerifySharedImageDestination) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	azcli := state.Get("azureclient").(client.AzureClientSet)
	ui := state.Get("ui").(packersdk.Ui)

	errorMessage := func(message string, parameters ...interface{}) multistep.StepAction {
		err := fmt.Errorf(message, parameters...)
		log.Printf("StepVerifySharedImageDestination.Run: error: %+v", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	imageURI := fmt.Sprintf("/subscriptions/%s/resourcegroup/%s/providers/Microsoft.Compute/galleries/%s/images/%s",
		azcli.SubscriptionID(),
		s.Image.ResourceGroup,
		s.Image.GalleryName,
		s.Image.ImageName,
	)

	ui.Say(fmt.Sprintf("Validating that shared image %s exists",
		imageURI))

	image, err := azcli.GalleryImagesClient().Get(ctx,
		s.Image.ResourceGroup,
		s.Image.GalleryName,
		s.Image.ImageName)

	if err != nil {
		return errorMessage("Error retrieving shared image %q: %+v ", imageURI, err)
	}

	if image.ID == nil || *image.ID == "" {
		return errorMessage("Error retrieving shared image %q: ID field in response is empty", imageURI)
	}
	if image.GalleryImageProperties == nil {
		return errorMessage("Could not retrieve shared image properties for image %q.", to.String(image.ID))
	}

	location := to.String(image.Location)

	log.Printf("StepVerifySharedImageDestination:Run: Image %q, Location: %q, HvGen: %q, osState: %q",
		to.String(image.ID),
		location,
		image.GalleryImageProperties.HyperVGeneration,
		image.GalleryImageProperties.OsState)

	if !strings.EqualFold(location, s.Location) {
		return errorMessage("Destination shared image resource %q is in a different location (%q) than this VM (%q). "+
			"Packer does not know how to handle that.",
			to.String(image.ID),
			location,
			s.Location)
	}

	if image.GalleryImageProperties.OsType != compute.Linux {
		return errorMessage("The shared image (%q) is not a Linux image (found %q). Currently only Linux images are supported.",
			to.String(image.ID),
			image.GalleryImageProperties.OsType)
	}

	ui.Say(fmt.Sprintf("Found image %s in location %s",
		to.String(image.ID),
		to.String(image.Location)))

	versions, err := azcli.GalleryImageVersionsClient().ListByGalleryImageComplete(ctx,
		s.Image.ResourceGroup,
		s.Image.GalleryName,
		s.Image.ImageName)

	if err != nil {
		return errorMessage("Could not ListByGalleryImageComplete group:%v gallery:%v image:%v",
			s.Image.ResourceGroup, s.Image.GalleryName, s.Image.ImageName)
	}

	for versions.NotDone() {
		version := versions.Value()

		if version.Name == nil {
			return errorMessage("Could not retrieve versions for image %q: unexpected nil name", to.String(image.ID))
		}
		if *version.Name == s.Image.ImageVersion {
			return errorMessage("Shared image version %q already exists for image %q.", s.Image.ImageVersion, to.String(image.ID))
		}

		err := versions.NextWithContext(ctx)
		if err != nil {
			return errorMessage("Could not retrieve versions for image %q: %+v", to.String(image.ID), err)
		}
	}

	return multistep.ActionContinue
}

func (*StepVerifySharedImageDestination) Cleanup(multistep.StateBag) {}
