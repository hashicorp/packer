package common

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

func getValidationSession() *ec2.EC2 {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	ec2conn := ec2.New(sess)
	return ec2conn
}

func listEC2Regions(ec2conn ec2iface.EC2API) []string {
	var regions []string
	resultRegions, _ := ec2conn.DescribeRegions(nil)
	for _, region := range resultRegions.Regions {
		regions = append(regions, *region.RegionName)
	}

	return regions
}

// ValidateRegion returns true if the supplied region is a valid AWS
// region and false if it's not.
func ValidateRegion(region string, ec2conn ec2iface.EC2API) bool {
	for _, valid := range listEC2Regions(ec2conn) {
		if region == valid {
			return true
		}
	}
	return false
}
