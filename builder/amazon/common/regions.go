package common

import (
	"fmt"
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
func ValidateRegion(region string, ec2conn ec2iface.EC2API) error {
	regions, err := listEC2Regions(ec2conn)
	if err != nil {
		return err
	}

	for _, valid := range regions {
		if region == valid {
			return nil
		}
	}

	return fmt.Errorf("Invalid region: %s", region)
}
