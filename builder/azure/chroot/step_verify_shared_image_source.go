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

var _ multistep.Step = &StepVerifySharedImageSource{}

// StepVerifySharedImageSource verifies that the shared image location matches the Location field in the step.
// Also verifies that the OS Type is Linux.
type StepVerifySharedImageSource struct {
	SharedImageID  string
	SubscriptionID string
	Location       string
}

// Run retrieves the image metadata from Azure and compares the location to Location. Verifies the OS Type.
func (s *StepVerifySharedImageSource) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	azcli := state.Get("azureclient").(client.AzureClientSet)
	ui := state.Get("ui").(packersdk.Ui)

	errorMessage := func(message string, parameters ...interface{}) multistep.StepAction {
		err := fmt.Errorf(message, parameters...)
		log.Printf("StepVerifySharedImageSource.Run: error: %+v", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	resource, err := client.ParseResourceID(s.SharedImageID)
	if err != nil {
		return errorMessage("Could not parse resource id %q: %w", s.SharedImageID, err)
	}

	if !strings.EqualFold(resource.Provider, "Microsoft.Compute") ||
		!strings.EqualFold(resource.ResourceType.String(), "galleries/images/versions") {
		return errorMessage("Resource id %q does not identify a shared image version, expected Microsoft.Compute/galleries/images/versions", s.SharedImageID)
	}

	ui.Say(fmt.Sprintf("Validating that shared image version %q exists",
		s.SharedImageID))

	version, err := azcli.GalleryImageVersionsClient().Get(ctx,
		resource.ResourceGroup,
		resource.ResourceName[0],
		resource.ResourceName[1],
		resource.ResourceName[2],
		"")

	if err != nil {
		return errorMessage("Error retrieving shared image version %q: %+v ", s.SharedImageID, err)
	}

	if version.ID == nil || *version.ID == "" {
		return errorMessage("Error retrieving shared image version %q: ID field in response is empty", s.SharedImageID)
	}

	if version.GalleryImageVersionProperties == nil ||
		version.GalleryImageVersionProperties.PublishingProfile == nil ||
		version.GalleryImageVersionProperties.PublishingProfile.TargetRegions == nil {
		return errorMessage("Could not retrieve shared image version properties for image %q.", s.SharedImageID)
	}

	targetLocations := make([]string, 0, len(*version.GalleryImageVersionProperties.PublishingProfile.TargetRegions))
	vmLocation := client.NormalizeLocation(s.Location)
	locationFound := false
	for _, tr := range *version.GalleryImageVersionProperties.PublishingProfile.TargetRegions {
		l := to.String(tr.Name)
		l = client.NormalizeLocation(l)
		targetLocations = append(targetLocations, l)
		if strings.EqualFold(vmLocation, l) {
			locationFound = true
			break
		}
	}
	if !locationFound {
		return errorMessage("Target locations %q for %q does not include VM location %q",
			targetLocations, s.SharedImageID, vmLocation)
	}

	imageResource, _ := resource.Parent()
	image, err := azcli.GalleryImagesClient().Get(ctx,
		resource.ResourceGroup,
		resource.ResourceName[0],
		resource.ResourceName[1])

	if err != nil {
		return errorMessage("Error retrieving shared image %q: %+v ", imageResource.String(), err)
	}

	if image.ID == nil || *image.ID == "" {
		return errorMessage("Error retrieving shared image %q: ID field in response is empty", imageResource.String())
	}

	if image.GalleryImageProperties == nil {
		return errorMessage("Could not retrieve shared image properties for image %q.", imageResource.String())
	}

	log.Printf("StepVerifySharedImageSource:Run: Image %q, HvGen: %q, osState: %q",
		to.String(image.ID),
		image.GalleryImageProperties.HyperVGeneration,
		image.GalleryImageProperties.OsState)

	if image.GalleryImageProperties.OsType != compute.Linux {
		return errorMessage("The shared image (%q) is not a Linux image (found %q). Currently only Linux images are supported.",
			to.String(image.ID),
			image.GalleryImageProperties.OsType)
	}

	ui.Say(fmt.Sprintf("Found image source image version %q, available in location %s",
		s.SharedImageID,
		s.Location))

	return multistep.ActionContinue
}

func (*StepVerifySharedImageSource) Cleanup(multistep.StateBag) {}
