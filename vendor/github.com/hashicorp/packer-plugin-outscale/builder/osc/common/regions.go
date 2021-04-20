package common

import (
	"context"
	"fmt"

	"github.com/antihax/optional"
	"github.com/outscale/osc-sdk-go/osc"
)

func listOSCRegions(oscconn *osc.RegionApiService) ([]string, error) {
	var regions []string
	resp, _, err := oscconn.ReadRegions(context.Background(), &osc.ReadRegionsOpts{
		ReadRegionsRequest: optional.NewInterface(osc.ReadRegionsRequest{}),
	})
	if err != nil {
		return []string{}, err
	}

	resultRegions := resp

	for _, region := range resultRegions.Regions {
		regions = append(regions, region.RegionName)
	}

	return regions, nil
}

// ValidateRegion returns true if the supplied region is a valid Outscale
// region and false if it's not.
func (c *AccessConfig) ValidateOSCRegion(regions ...string) error {
	oscconn := c.NewOSCClient()

	validRegions, err := listOSCRegions(oscconn.RegionApi)
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
