package common

import (
	"fmt"

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
	AmiFilters         DynamicAmiOptions
}

// Build a slice of AMI filter options from the filters provided.
func buildAmiFilters(input map[*string]*string) []*ec2.Filter {
	var filters []*ec2.Filter
	for k, v := range input {
		/*m := v.(map[string]interface{})
		  var filterValues []*string
		  for _, e := range m["values"].([]interface{}) {
		      filterValues = append(filterValues, aws.String(e.(string)))
		  }*/
		filters = append(filters, &ec2.Filter{
			Name:   k,
			Values: []*string{v},
		})
	}
	return filters
}

func (s *StepSourceAMIInfo) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	params := &ec2.DescribeImagesInput{}
	params.Filters = buildAmiFilters(s.AmiFilters.Filters)
	params.Owners = s.AmiFilters.Owners
	ui.Say(fmt.Sprintf("Using AMI Filters %v", params))
	imageResp, err := ec2conn.DescribeImages(params)
	//ui.Say(fmt.Sprintf("Inspecting the source AMI (%s)...", s.SourceAmi))
	//imageResp, err := ec2conn.DescribeImages(&ec2.DescribeImagesInput{ImageIds: []*string{&s.SourceAmi}})
	if err != nil {
		err := fmt.Errorf("Error querying AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(imageResp.Images) == 0 {
		err := fmt.Errorf("Source AMI '%s' was not found!", s.SourceAmi)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	image := imageResp.Images[0]
	ui.Say(fmt.Sprintf("Got Image %v", image))

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
