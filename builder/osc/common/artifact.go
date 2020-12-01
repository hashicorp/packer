package common

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/antihax/optional"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/outscale/osc-sdk-go/osc"
)

// Artifact is an artifact implementation that contains built OMIs.
type Artifact struct {
	// A map of regions to OMI IDs.
	Omis map[string]string

	// BuilderId is the unique ID for the builder that created this OMI
	BuilderIdValue string

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

func (a *Artifact) BuilderId() string {
	return a.BuilderIdValue
}

func (*Artifact) Files() []string {
	// We have no files
	return nil
}

func (a *Artifact) Id() string {
	parts := make([]string, 0, len(a.Omis))
	for region, amiId := range a.Omis {
		parts = append(parts, fmt.Sprintf("%s:%s", region, amiId))
	}

	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func (a *Artifact) String() string {
	amiStrings := make([]string, 0, len(a.Omis))
	for region, id := range a.Omis {
		single := fmt.Sprintf("%s: %s", region, id)
		amiStrings = append(amiStrings, single)
	}

	sort.Strings(amiStrings)
	return fmt.Sprintf("OMIs were created:\n%s\n", strings.Join(amiStrings, "\n"))
}

func (a *Artifact) State(name string) interface{} {
	if _, ok := a.StateData[name]; ok {
		return a.StateData[name]
	}

	switch name {
	case "atlas.artifact.metadata":
		return a.stateAtlasMetadata()
	default:
		return nil
	}
}

func (a *Artifact) Destroy() error {
	errors := make([]error, 0)

	config := a.State("accessConfig").(*AccessConfig)

	for region, imageId := range a.Omis {
		log.Printf("Deregistering image ID (%s) from region (%s)", imageId, region)

		regionConn := config.NewOSCClientByRegion(region)

		// Get image metadata
		imageResp, _, err := regionConn.ImageApi.ReadImages(context.Background(), &osc.ReadImagesOpts{
			ReadImagesRequest: optional.NewInterface(osc.ReadImagesRequest{
				Filters: osc.FiltersImage{
					ImageIds: []string{imageId},
				},
			}),
		})
		if err != nil {
			errors = append(errors, err)
		}
		if len(imageResp.Images) == 0 {
			err := fmt.Errorf("Error retrieving details for OMI (%s), no images found", imageId)
			errors = append(errors, err)
		}

		// Deregister ami
		input := osc.DeleteImageRequest{
			ImageId: imageId,
		}
		if _, _, err := regionConn.ImageApi.DeleteImage(context.Background(), &osc.DeleteImageOpts{
			DeleteImageRequest: optional.NewInterface(input),
		}); err != nil {
			errors = append(errors, err)
		}

		// TODO: Delete the snapshots associated with an OMI too
	}

	if len(errors) > 0 {
		if len(errors) == 1 {
			return errors[0]
		} else {
			return &packersdk.MultiError{Errors: errors}
		}
	}

	return nil
}

func (a *Artifact) stateAtlasMetadata() interface{} {
	metadata := make(map[string]string)
	for region, imageId := range a.Omis {
		k := fmt.Sprintf("region.%s", region)
		metadata[k] = imageId
	}

	return metadata
}
