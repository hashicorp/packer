package common

import (
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Build a slice of EC2 (AMI/Subnet/VPC) filter options from the filters provided.
func buildEc2Filters(input map[string]string) []*ec2.Filter {
	var filters []*ec2.Filter
	for k, v := range input {
		a := k
		b := v
		filters = append(filters, &ec2.Filter{
			Name:   &a,
			Values: []*string{&b},
		})
	}
	return filters
}
