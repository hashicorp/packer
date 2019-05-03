package common

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

func boolPointer(tf bool) *bool {
	return &tf
}

// Define a mock struct to be used in unit tests for common aws steps.
type mockEC2Conn struct {
	ec2iface.EC2API
	Config *aws.Config

	// Counters to figure out what code path was taken
	copyImageCount       int
	describeImagesCount  int
	deregisterImageCount int
	deleteSnapshotCount  int
	waitCount            int
}

func (m *mockEC2Conn) CopyImage(copyInput *ec2.CopyImageInput) (*ec2.CopyImageOutput, error) {
	m.copyImageCount++
	copiedImage := fmt.Sprintf("%s-copied-%d", *copyInput.SourceImageId, m.copyImageCount)
	output := &ec2.CopyImageOutput{
		ImageId: &copiedImage,
	}
	return output, nil
}

// functions we have to create mock responses for in order for test to run
func (m *mockEC2Conn) DescribeImages(*ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	m.describeImagesCount++
	output := &ec2.DescribeImagesOutput{
		Images: []*ec2.Image{{}},
	}
	return output, nil
}

func (m *mockEC2Conn) DeregisterImage(*ec2.DeregisterImageInput) (*ec2.DeregisterImageOutput, error) {
	m.deregisterImageCount++
	output := &ec2.DeregisterImageOutput{}
	return output, nil
}

func (m *mockEC2Conn) DeleteSnapshot(*ec2.DeleteSnapshotInput) (*ec2.DeleteSnapshotOutput, error) {
	m.deleteSnapshotCount++
	output := &ec2.DeleteSnapshotOutput{}
	return output, nil
}

func (m *mockEC2Conn) WaitUntilImageAvailableWithContext(aws.Context, *ec2.DescribeImagesInput, ...request.WaiterOption) error {
	m.waitCount++
	return nil
}

func getMockConn(config *AccessConfig, target string) (ec2iface.EC2API, error) {
	mockConn := &mockEC2Conn{
		Config: aws.NewConfig(),
	}

	return mockConn, nil
}

// Create statebag for running test
func tState() multistep.StateBag {
	state := new(multistep.BasicStateBag)
	state.Put("ui", &packer.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	state.Put("amis", map[string]string{"us-east-1": "ami-12345"})
	state.Put("snapshots", map[string][]string{"us-east-1": {"snap-0012345"}})
	conn, _ := getMockConn(&AccessConfig{}, "us-east-2")
	state.Put("ec2", conn)
	return state
}

func TestStepAmiRegionCopy_nil_encryption(t *testing.T) {
	// create step
	stepAMIRegionCopy := StepAMIRegionCopy{
		AccessConfig:      testAccessConfig(),
		Regions:           make([]string, 0),
		AMIKmsKeyId:       "",
		RegionKeyIds:      make(map[string]string),
		EncryptBootVolume: nil,
		Name:              "fake-ami-name",
		OriginalRegion:    "us-east-1",
	}
	// mock out the region connection code
	stepAMIRegionCopy.getRegionConn = getMockConn

	state := tState()
	stepAMIRegionCopy.Run(context.Background(), state)

	if stepAMIRegionCopy.toDelete != "" {
		t.Fatalf("Shouldn't delete original AMI if not encrypted")
	}
	if len(stepAMIRegionCopy.Regions) > 0 {
		t.Fatalf("Shouldn't have added original ami to original region")
	}
}

func TestStepAmiRegionCopy_false_encryption(t *testing.T) {
	// create step
	stepAMIRegionCopy := StepAMIRegionCopy{
		AccessConfig:      testAccessConfig(),
		Regions:           make([]string, 0),
		AMIKmsKeyId:       "",
		RegionKeyIds:      make(map[string]string),
		EncryptBootVolume: boolPointer(false),
		Name:              "fake-ami-name",
		OriginalRegion:    "us-east-1",
	}
	// mock out the region connection code
	stepAMIRegionCopy.getRegionConn = getMockConn

	state := tState()
	stepAMIRegionCopy.Run(context.Background(), state)

	if stepAMIRegionCopy.toDelete != "" {
		t.Fatalf("Shouldn't delete original AMI if not encrypted")
	}
	if len(stepAMIRegionCopy.Regions) > 0 {
		t.Fatalf("Shouldn't have added original ami to Regions")
	}
}

func TestStepAmiRegionCopy_true_encryption(t *testing.T) {
	// create step
	stepAMIRegionCopy := StepAMIRegionCopy{
		AccessConfig:      testAccessConfig(),
		Regions:           make([]string, 0),
		AMIKmsKeyId:       "",
		RegionKeyIds:      make(map[string]string),
		EncryptBootVolume: boolPointer(true),
		Name:              "fake-ami-name",
		OriginalRegion:    "us-east-1",
	}
	// mock out the region connection code
	stepAMIRegionCopy.getRegionConn = getMockConn

	state := tState()
	stepAMIRegionCopy.Run(context.Background(), state)

	if stepAMIRegionCopy.toDelete == "" {
		t.Fatalf("Should delete original AMI if encrypted=true")
	}
	if len(stepAMIRegionCopy.Regions) == 0 {
		t.Fatalf("Should have added original ami to Regions")
	}
}
