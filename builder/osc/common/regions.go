package common

import (
	"fmt"

	"github.com/outscale/osc-go/oapi"
)

func listOAPIRegions(oapiconn oapi.OAPIClient) ([]string, error) {
	var regions []string
	resp, err := oapiconn.POST_ReadRegions(oapi.ReadRegionsRequest{})
	if resp.OK == nil || err != nil {
		return []string{}, err
	}

	resultRegions := resp.OK

	for _, region := range resultRegions.Regions {
		regions = append(regions, region.RegionName)
	}

	return regions, nil
}

// ValidateRegion returns true if the supplied region is a valid Outscale
// region and false if it's not.
func (c *AccessConfig) ValidateRegion(regions ...string) error {
	oapiconn, err := c.NewOAPIConnection()
	if err != nil {
		return err
	}

	validRegions, err := listOAPIRegions(oapiconn)
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
