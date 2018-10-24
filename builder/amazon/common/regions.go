package common

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

func listEC2Regions(ec2conn ec2iface.EC2API) ([]string, error) {
	var regions []string
	resultRegions, err := ec2conn.DescribeRegions(nil)
	if err != nil {
		return []string{}, err
	}
	for _, region := range resultRegions.Regions {
		regions = append(regions, *region.RegionName)
	}

	return regions, nil
}

// ValidateRegion returns true if the supplied region is a valid AWS
// region and false if it's not.
func (c *AccessConfig) ValidateRegion(regions ...string) error {
	ec2conn, err := c.NewEC2Connection()
	if err != nil {
		return err
	}

	validRegions, err := listEC2Regions(ec2conn)
	if err != nil {
		return err
	}

	var invalidRegions []string
	for _, region := range regions {
		found := false
		for _, validRegion := range validRegions {
			if region == validRegion {
				found = true
				break
			}
		}
		if !found {
			invalidRegions = append(invalidRegions, region)
		}
	}

	if len(invalidRegions) > 0 {
		return fmt.Errorf("Invalid region(s): %v", invalidRegions)
	}
	return nil
}
