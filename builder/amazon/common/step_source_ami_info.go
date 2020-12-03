package common

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	confighelper "github.com/hashicorp/packer/packer-plugin-sdk/template/config"
)

// StepSourceAMIInfo extracts critical information from the source AMI
// that is used throughout the AMI creation process.
//
// Produces:
//   source_image *ec2.Image - the source AMI info
type StepSourceAMIInfo struct {
	SourceAmi                string
	EnableAMISriovNetSupport bool
	EnableAMIENASupport      confighelper.Trilean
	AMIVirtType              string
	AmiFilters               AmiFilterOptions
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

func (s *StepSourceAMIInfo) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packersdk.Ui)

	params := &ec2.DescribeImagesInput{}

	if s.SourceAmi != "" {
		params.ImageIds = []*string{&s.SourceAmi}
	}

	// We have filters to apply
	if len(s.AmiFilters.Filters) > 0 {
		params.Filters = buildEc2Filters(s.AmiFilters.Filters)
	}
	if len(s.AmiFilters.Owners) > 0 {
		params.Owners = s.AmiFilters.GetOwners()
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

	if len(imageResp.Images) > 1 && !s.AmiFilters.MostRecent {
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

	ui.Message(fmt.Sprintf("Found Image ID: %s", *image.ImageId))

	// Enhanced Networking can only be enabled on HVM AMIs.
	// See http://goo.gl/icuXh5
	if s.EnableAMIENASupport.True() || s.EnableAMISriovNetSupport {
		err = s.canEnableEnhancedNetworking(image)
		if err != nil {
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	state.Put("source_image", image)
	return multistep.ActionContinue
}

func (s *StepSourceAMIInfo) Cleanup(multistep.StateBag) {}

func (s *StepSourceAMIInfo) canEnableEnhancedNetworking(image *ec2.Image) error {
	if s.AMIVirtType == "hvm" {
		return nil
	}
	if s.AMIVirtType != "" {
		return fmt.Errorf("Cannot enable enhanced networking, AMIVirtType '%s' is not HVM", s.AMIVirtType)
	}
	if *image.VirtualizationType != "hvm" {
		return fmt.Errorf("Cannot enable enhanced networking, source AMI '%s' is not HVM", s.SourceAmi)
	}
	return nil
}
