package common

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

//DescribeVpcs mocks an ec2.DescribeVpcsOutput for a given input
func (m *mockEC2Conn) DescribeVpcs(input *ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
	m.lock.Lock()
	m.copyImageCount++
	m.lock.Unlock()

	if input == nil || len(input.VpcIds) == 0 {
		return nil, fmt.Errorf("oops looks like we need more input")
	}

	var isDefault bool
	vpcID := aws.StringValue(input.VpcIds[0])

	//only one default VPC per region
	if strings.Contains("vpc-default-id", vpcID) {
		isDefault = true
	}

	output := &ec2.DescribeVpcsOutput{
		Vpcs: []*ec2.Vpc{
			&ec2.Vpc{IsDefault: aws.Bool(isDefault),
				VpcId: aws.String(vpcID),
			},
		},
	}
	return output, nil
}

func TestStepPreValidate_checkVpc(t *testing.T) {
	tt := []struct {
		name          string
		step          StepPreValidate
		errorExpected bool
	}{
		{"DefaultVpc", StepPreValidate{VpcId: "vpc-default-id"}, false},
		{"NonDefaultVpcNoSubnet", StepPreValidate{VpcId: "vpc-1234567890"}, true},
		{"NonDefaultVpcWithSubnet", StepPreValidate{VpcId: "vpc-1234567890", SubnetId: "subnet-1234567890"}, false},
		{"SubnetWithNoVpc", StepPreValidate{SubnetId: "subnet-1234567890"}, false},
		{"NoVpcInformation", StepPreValidate{}, false},
	}

	mockConn, err := getMockConn(nil, "")
	if err != nil {
		t.Fatal("unable to get a mock connection")
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.step.checkVpc(mockConn)

			if tc.errorExpected && err == nil {
				t.Errorf("expected a validation error for %q but got %q", tc.name, err)
			}

			if !tc.errorExpected && err != nil {
				t.Errorf("expected a validation to pass for %q but got %q", tc.name, err)
			}
		})
	}

}
