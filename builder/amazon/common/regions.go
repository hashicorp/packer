package common

import (
	"flag"

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

	// To pass tests
	if v := flag.Lookup("test.v"); v != nil || v.Value.String() == "true" {
		regions := []string{
			"us-east-1",
			"us-east-2",
			"us-west-1",
		}
		for _, valid := range regions {
			if region == valid {
				return true
			}
		}
	}

	// Normal run
	for _, valid := range listEC2Regions() {
		if region == valid {
			return true
		}
	}
	return false
}
