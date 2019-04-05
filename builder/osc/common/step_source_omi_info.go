package common

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/outscale/osc-go/oapi"
)

// StepSourceOMIInfo extracts critical information from the source OMI
// that is used throughout the OMI creation process.
//
// Produces:
//   source_image *oapi.Image - the source OMI info
type StepSourceOMIInfo struct {
	SourceOmi   string
	OMIVirtType string
	OmiFilters  OmiFilterOptions
}

type imageSort []oapi.Image

func (a imageSort) Len() int      { return len(a) }
func (a imageSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a imageSort) Less(i, j int) bool {
	itime, _ := time.Parse(time.RFC3339, a[i].CreationDate)
	jtime, _ := time.Parse(time.RFC3339, a[j].CreationDate)
	return itime.Unix() < jtime.Unix()
}

// Returns the most recent OMI out of a slice of images.
func mostRecentOmi(images []oapi.Image) oapi.Image {
	sortedImages := images
	sort.Sort(imageSort(sortedImages))
	return sortedImages[len(sortedImages)-1]
}

func (s *StepSourceOMIInfo) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	oapiconn := state.Get("oapi").(*oapi.Client)
	ui := state.Get("ui").(packer.Ui)

	params := oapi.ReadImagesRequest{
		Filters: oapi.FiltersImage{},
	}

	if s.SourceOmi != "" {
		params.Filters.ImageIds = []string{s.SourceOmi}
	}

	// We have filters to apply
	if len(s.OmiFilters.Filters) > 0 {
		params.Filters = buildOMIFilters(s.OmiFilters.Filters)
	}
	//TODO:Check if AccountIds correspond to Owners.
	if len(s.OmiFilters.Owners) > 0 {
		params.Filters.AccountIds = s.OmiFilters.Owners
	}

	log.Printf("Using OMI Filters %#v", params)
	imageResp, err := oapiconn.POST_ReadImages(params)
	if err != nil {
		err := fmt.Errorf("Error querying OMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(imageResp.OK.Images) == 0 {
		err := fmt.Errorf("No OMI was found matching filters: %#v", params)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(imageResp.OK.Images) > 1 && !s.OmiFilters.MostRecent {
		err := fmt.Errorf("Your query returned more than one result. Please try a more specific search, or set most_recent to true.")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	var image oapi.Image
	if s.OmiFilters.MostRecent {
		image = mostRecentOmi(imageResp.OK.Images)
	} else {
		image = imageResp.OK.Images[0]
	}

	ui.Message(fmt.Sprintf("Found Image ID: %s", image.ImageId))

	state.Put("source_image", image)
	return multistep.ActionContinue
}

func (s *StepSourceOMIInfo) Cleanup(multistep.StateBag) {}
