package common

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/hashicorp/packer/packer"
	"github.com/outscale/osc-go/oapi"
)

// Artifact is an artifact implementation that contains built OMIs.
type Artifact struct {
	// A map of regions to OMI IDs.
	Omis map[string]string

	// BuilderId is the unique ID for the builder that created this OMI
	BuilderIdValue string

	// OAPI connection for performing API stuff.
	Config *oapi.Config
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
	switch name {
	case "atlas.artifact.metadata":
		return a.stateAtlasMetadata()
	default:
		return nil
	}
}

func (a *Artifact) Destroy() error {
	errors := make([]error, 0)

	for region, imageId := range a.Omis {
		log.Printf("Deregistering image ID (%s) from region (%s)", imageId, region)

		newConfig := &oapi.Config{
			UserAgent: a.Config.UserAgent,
			AccessKey: a.Config.AccessKey,
			SecretKey: a.Config.SecretKey,
			Service:   a.Config.Service,
			Region:    region, //New region
			URL:       a.Config.URL,
		}

		log.Printf("[DEBUG] New Client config %+v", newConfig)

		skipClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		regionConn := oapi.NewClient(newConfig, skipClient)

		// Get image metadata
		imageResp, err := regionConn.POST_ReadImages(oapi.ReadImagesRequest{
			Filters: oapi.FiltersImage{
				ImageIds: []string{imageId},
			},
		})
		if err != nil {
			errors = append(errors, err)
		}
		if len(imageResp.OK.Images) == 0 {
			err := fmt.Errorf("Error retrieving details for OMI (%s), no images found", imageId)
			errors = append(errors, err)
		}

		// Deregister ami
		input := oapi.DeleteImageRequest{
			ImageId: imageId,
		}
		if _, err := regionConn.POST_DeleteImage(input); err != nil {
			errors = append(errors, err)
		}

		// TODO: Delete the snapshots associated with an OMI too
	}

	if len(errors) > 0 {
		if len(errors) == 1 {
			return errors[0]
		} else {
			return &packer.MultiError{Errors: errors}
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
