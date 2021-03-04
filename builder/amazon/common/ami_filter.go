//go:generate struct-markdown
package common

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
)

type AmiFilterOptions struct {
	// Filters used to select an AMI. Any filter described in the docs for
	// [DescribeImages](http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeImages.html)
	// is valid.
	Filters map[string]string `mapstructure:"filters"`
	// Filters the images by their owner. You
	// may specify one or more AWS account IDs, "self" (which will use the
	// account whose credentials you are using to run Packer), or an AWS owner
	// alias: for example, `amazon`, `aws-marketplace`, or `microsoft`. This
	// option is required for security reasons.
	Owners []string `mapstructure:"owners"`
	// Selects the newest created image when true.
	// This is most useful for selecting a daily distro build.
	MostRecent bool `mapstructure:"most_recent"`
}

func (d *AmiFilterOptions) GetOwners() []*string {
	res := make([]*string, 0, len(d.Owners))
	for _, owner := range d.Owners {
		i := owner
		res = append(res, &i)
	}
	return res
}

func (d *AmiFilterOptions) Empty() bool {
	return len(d.Owners) == 0 && len(d.Filters) == 0
}

func (d *AmiFilterOptions) NoOwner() bool {
	return len(d.Owners) == 0
}

func (d *AmiFilterOptions) GetFilteredImage(params *ec2.DescribeImagesInput, ec2conn *ec2.EC2) (*ec2.Image, error) {
	// We have filters to apply
	if len(d.Filters) > 0 {
		params.Filters = buildEc2Filters(d.Filters)
	}
	if len(d.Owners) > 0 {
		params.Owners = d.GetOwners()
	}

	log.Printf("Using AMI Filters %v", params)
	req, imageResp := ec2conn.DescribeImagesRequest(params)
	req.RetryCount = 11

	err := req.Send()
	if err != nil {
		err := fmt.Errorf("Error querying AMI: %s", err)
		return nil, err
	}

	if len(imageResp.Images) == 0 {
		err := fmt.Errorf("No AMI was found matching filters: %v", params)
		return nil, err
	}

	if len(imageResp.Images) > 1 && !d.MostRecent {
		err := fmt.Errorf("Your query returned more than one result. Please try a more specific search, or set most_recent to true.")
		return nil, err
	}

	var image *ec2.Image
	if d.MostRecent {
		image = mostRecentAmi(imageResp.Images)
	} else {
		image = imageResp.Images[0]
	}
	return image, nil
}
