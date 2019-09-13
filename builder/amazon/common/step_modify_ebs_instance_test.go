package common

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	helperconfig "github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
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

// Create statebag for running test
func fakeModifyEBSBackedInstanceState() multistep.StateBag {
	state := new(multistep.BasicStateBag)
	state.Put("ui", &packer.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	state.Put("instance", "i-12345")
	return state
}

func StepModifyEBSBackedInstance_EnableAMIENASupport(t *testing.T) {
	// Value is unset, so we shouldn't modify
	stepModifyEBSBackedInstance := StepModifyEBSBackedInstance{
		EnableAMIENASupport:      helperconfig.TriUnset,
		EnableAMISriovNetSupport: false,
	}

	// mock out the region connection code
	mockConn := &mockEC2Conn_ModifyEBS{
		Config: aws.NewConfig(),
	}

	state := fakeModifyEBSBackedInstanceState()
	state.Put("ec2", mockConn)
	stepModifyEBSBackedInstance.Run(context.Background(), state)

	if mockConn.modifyImageAttrCount > 0 {
		t.Fatalf("Should not have modified image since EnableAMIENASupport is unset")
	}

	// Value is true, so we should modify
	stepModifyEBSBackedInstance = StepModifyEBSBackedInstance{
		EnableAMIENASupport:      helperconfig.TriTrue,
		EnableAMISriovNetSupport: false,
	}

	// mock out the region connection code
	mockConn = &mockEC2Conn_ModifyEBS{
		Config: aws.NewConfig(),
	}

	state = fakeModifyEBSBackedInstanceState()
	state.Put("ec2", mockConn)
	stepModifyEBSBackedInstance.Run(context.Background(), state)

	if mockConn.modifyImageAttrCount != 1 {
		t.Fatalf("Should have modified image, since EnableAMIENASupport is true")
	}

	// Value is false, so we should modify
	stepModifyEBSBackedInstance = StepModifyEBSBackedInstance{
		EnableAMIENASupport:      helperconfig.TriFalse,
		EnableAMISriovNetSupport: false,
	}

	// mock out the region connection code
	mockConn = &mockEC2Conn_ModifyEBS{
		Config: aws.NewConfig(),
	}

	state = fakeModifyEBSBackedInstanceState()
	state.Put("ec2", mockConn)
	stepModifyEBSBackedInstance.Run(context.Background(), state)

	if mockConn.modifyImageAttrCount != 1 {
		t.Fatalf("Should have modified image, since EnableAMIENASupport is true")
	}
}
