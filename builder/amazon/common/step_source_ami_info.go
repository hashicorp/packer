package common

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// StepSourceAMIInfo extracts critical information from the source AMI
// that is used throughout the AMI creation process.
//
// Produces:
//   source_image *ec2.Image - the source AMI info
type StepSourceAMIInfo struct {
	SourceAmi          string
	EnhancedNetworking bool
	AmiFilters         AmiFilterOptions
}

// Build a slice of AMI filter options from the filters provided.
func buildAmiFilters(input map[*string]*string) []*ec2.Filter {
	var filters []*ec2.Filter
	for k, v := range input {
		filters = append(filters, &ec2.Filter{
			Name:   k,
			Values: []*string{v},
		})
	}
	return filters
}

type imageSort []*ec2.Image

func (a imageSort) Len() int      { return len(a) }
func (a imageSort) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a imageSort) Less(i, j int) bool {
	itime, _ := time.Parse(time.RFC3339, *a[i].CreationDate)
	jtime, _ := time.Parse(time.RFC3339, *a[j].CreationDate)
	return itime.Unix() < jtime.Unix()
}

// Returns the most recent AMI out of a slice of images.
func mostRecentAmi(images []*ec2.Image) *ec2.Image {
	sortedImages := images
	sort.Sort(imageSort(sortedImages))
	return sortedImages[len(sortedImages)-1]
}

func (s *StepSourceAMIInfo) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	params := &ec2.DescribeImagesInput{}

	if s.SourceAmi != "" {
		params.ImageIds = []*string{&s.SourceAmi}
	}

	// We have filters to apply
	if len(s.AmiFilters.Filters) > 0 {
		params.Filters = buildAmiFilters(s.AmiFilters.Filters)
	}

	log.Printf("Using AMI Filters %v", params)
	imageResp, err := ec2conn.DescribeImages(params)
	if err != nil {
		err := fmt.Errorf("Error querying AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(imageResp.Images) == 0 {
		err := fmt.Errorf("No AMI was found matching filters: %v", params)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(imageResp.Images) > 1 && s.AmiFilters.MostRecent == false {
		err := fmt.Errorf("Your query returned more than one result. Please try a more specific search, or set most_recent to true.")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	var image *ec2.Image
	if s.AmiFilters.MostRecent {
		image = mostRecentAmi(imageResp.Images)
	} else {
		image = imageResp.Images[0]
	}

	log.Printf(fmt.Sprintf("Got Image %v", image))

	// Enhanced Networking (SriovNetSupport) can only be enabled on HVM AMIs.
	// See http://goo.gl/icuXh5
	if s.EnhancedNetworking && *image.VirtualizationType != "hvm" {
		err := fmt.Errorf("Cannot enable enhanced networking, source AMI '%s' is not HVM", s.SourceAmi)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("source_image", image)
	return multistep.ActionContinue
}

func (s *StepSourceAMIInfo) Cleanup(multistep.StateBag) {}
