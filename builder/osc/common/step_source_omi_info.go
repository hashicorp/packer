package common

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/antihax/optional"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/outscale/osc-sdk-go/osc"
)

// StepSourceOMIInfo extracts critical information from the source OMI
// that is used throughout the OMI creation process.
//
// Produces:
//   source_image *osc.Image - the source OMI info
type StepSourceOMIInfo struct {
	SourceOmi   string
	OMIVirtType string
	OmiFilters  OmiFilterOptions
}

type imageOscSort []osc.Image

func (a imageOscSort) Len() int      { return len(a) }
func (a imageOscSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a imageOscSort) Less(i, j int) bool {
	itime, _ := time.Parse(time.RFC3339, a[i].CreationDate)
	jtime, _ := time.Parse(time.RFC3339, a[j].CreationDate)
	return itime.Unix() < jtime.Unix()
}

// Returns the most recent OMI out of a slice of images.
func mostRecentOscOmi(images []osc.Image) osc.Image {
	sortedImages := images
	sort.Sort(imageOscSort(sortedImages))
	return sortedImages[len(sortedImages)-1]
}

func (s *StepSourceOMIInfo) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	oscconn := state.Get("osc").(*osc.APIClient)
	ui := state.Get("ui").(packersdk.Ui)

	params := osc.ReadImagesRequest{
		Filters: osc.FiltersImage{},
	}

	if s.SourceOmi != "" {
		params.Filters.ImageIds = []string{s.SourceOmi}
	}

	// We have filters to apply
	if len(s.OmiFilters.Filters) > 0 {
		params.Filters = buildOSCOMIFilters(s.OmiFilters.Filters)
	}
	//TODO:Check if AccountIds correspond to Owners.
	if len(s.OmiFilters.Owners) > 0 {
		params.Filters.AccountIds = s.OmiFilters.Owners
	}

	log.Printf("Using OMI Filters %#v", params)
	imageResp, _, err := oscconn.ImageApi.ReadImages(context.Background(), &osc.ReadImagesOpts{
		ReadImagesRequest: optional.NewInterface(params),
	})
	if err != nil {
		err := fmt.Errorf("Error querying OMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(imageResp.Images) == 0 {
		err := fmt.Errorf("No OMI was found matching filters: %#v", params)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(imageResp.Images) > 1 && !s.OmiFilters.MostRecent {
		err := fmt.Errorf("your query returned more than one result. Please try a more specific search, or set most_recent to true")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	var image osc.Image
	if s.OmiFilters.MostRecent {
		image = mostRecentOscOmi(imageResp.Images)
	} else {
		image = imageResp.Images[0]
	}

	ui.Message(fmt.Sprintf("Found Image ID: %s", image.ImageId))

	state.Put("source_image", image)
	return multistep.ActionContinue
}

func (s *StepSourceOMIInfo) Cleanup(multistep.StateBag) {}
