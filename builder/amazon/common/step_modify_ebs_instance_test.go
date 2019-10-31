package common

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

// Define a mock struct to be used in unit tests for common aws steps.
type mockEC2Conn_ModifyEBS struct {
	ec2iface.EC2API
	Config *aws.Config

	// Counters to figure out what code path was taken
	shouldError          bool
	modifyImageAttrCount int
}

func (m *mockEC2Conn_ModifyEBS) ModifyInstanceAttribute(modifyInput *ec2.ModifyInstanceAttributeInput) (*ec2.ModifyInstanceAttributeOutput, error) {
	m.modifyImageAttrCount++
	// don't need to define output since we always discard it anyway.
	output := &ec2.ModifyInstanceAttributeOutput{}
	if m.shouldError {
		return output, fmt.Errorf("fake ModifyInstanceAttribute error")
	}
	return output, nil
}
