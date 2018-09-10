package common

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func listEC2Regions() []string {
	var regions []string
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	ec2conn := ec2.New(sess)
	resultRegions, _ := ec2conn.DescribeRegions(nil)
	for _, region := range resultRegions.Regions {
		regions = append(regions, *region.RegionName)
	}

	return regions
}

// ValidateRegion returns true if the supplied region is a valid AWS
// region and false if it's not.
func ValidateRegion(region string) bool {
	// Normal run
	for _, valid := range listEC2Regions() {
		if region == valid {
			return true
		}
	}
	return false
}
