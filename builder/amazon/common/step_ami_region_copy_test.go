package common

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
)

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

	lock sync.Mutex
}

func (m *mockEC2Conn) CopyImage(copyInput *ec2.CopyImageInput) (*ec2.CopyImageOutput, error) {
	m.lock.Lock()
	m.copyImageCount++
	m.lock.Unlock()
	copiedImage := fmt.Sprintf("%s-copied-%d", *copyInput.SourceImageId, m.copyImageCount)
	output := &ec2.CopyImageOutput{
		ImageId: &copiedImage,
	}
	return output, nil
}

// functions we have to create mock responses for in order for test to run
func (m *mockEC2Conn) DescribeImages(*ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	m.lock.Lock()
	m.describeImagesCount++
	m.lock.Unlock()
	output := &ec2.DescribeImagesOutput{
		Images: []*ec2.Image{{}},
	}
	return output, nil
}

func (m *mockEC2Conn) DeregisterImage(*ec2.DeregisterImageInput) (*ec2.DeregisterImageOutput, error) {
	m.lock.Lock()
	m.deregisterImageCount++
	m.lock.Unlock()
	output := &ec2.DeregisterImageOutput{}
	return output, nil
}

func (m *mockEC2Conn) DeleteSnapshot(*ec2.DeleteSnapshotInput) (*ec2.DeleteSnapshotOutput, error) {
	m.lock.Lock()
	m.deleteSnapshotCount++
	m.lock.Unlock()
	output := &ec2.DeleteSnapshotOutput{}
	return output, nil
}

func (m *mockEC2Conn) WaitUntilImageAvailableWithContext(aws.Context, *ec2.DescribeImagesInput, ...request.WaiterOption) error {
	m.lock.Lock()
	m.waitCount++
	m.lock.Unlock()
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
	state.Put("ui", &packersdk.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	state.Put("amis", map[string]string{"us-east-1": "ami-12345"})
	state.Put("snapshots", map[string][]string{"us-east-1": {"snap-0012345"}})
	conn, _ := getMockConn(&AccessConfig{}, "us-east-2")
	state.Put("ec2", conn)
	return state
}

func TestStepAMIRegionCopy_duplicates(t *testing.T) {
	// ------------------------------------------------------------------------
	// Test that if the original region is added to both Regions and Region,
	// the ami is only copied once (with encryption).
	// ------------------------------------------------------------------------

	stepAMIRegionCopy := StepAMIRegionCopy{
		AccessConfig: testAccessConfig(),
		Regions:      []string{"us-east-1"},
		AMIKmsKeyId:  "12345",
		// Original region key in regionkeyids is different than in amikmskeyid
		RegionKeyIds:      map[string]string{"us-east-1": "12345"},
		EncryptBootVolume: config.TriTrue,
		Name:              "fake-ami-name",
		OriginalRegion:    "us-east-1",
	}
	// mock out the region connection code
	stepAMIRegionCopy.getRegionConn = getMockConn

	state := tState()
	state.Put("intermediary_image", true)
	stepAMIRegionCopy.Run(context.Background(), state)

	if len(stepAMIRegionCopy.Regions) != 1 {
		t.Fatalf("Should have added original ami to Regions one time only")
	}

	// ------------------------------------------------------------------------
	// Both Region and Regions set, but no encryption - shouldn't copy anything
	// ------------------------------------------------------------------------

	// the ami is only copied once.
	stepAMIRegionCopy = StepAMIRegionCopy{
		AccessConfig:   testAccessConfig(),
		Regions:        []string{"us-east-1"},
		Name:           "fake-ami-name",
		OriginalRegion: "us-east-1",
	}
	// mock out the region connection code
	state.Put("intermediary_image", false)
	stepAMIRegionCopy.getRegionConn = getMockConn
	stepAMIRegionCopy.Run(context.Background(), state)

	if len(stepAMIRegionCopy.Regions) != 0 {
		t.Fatalf("Should not have added original ami to Regions; not encrypting")
	}

	// ------------------------------------------------------------------------
	// Both Region and Regions set, but no encryption - shouldn't copy anything,
	// this tests false as opposed to nil value above.
	// ------------------------------------------------------------------------

	// the ami is only copied once.
	stepAMIRegionCopy = StepAMIRegionCopy{
		AccessConfig:      testAccessConfig(),
		Regions:           []string{"us-east-1"},
		EncryptBootVolume: config.TriFalse,
		Name:              "fake-ami-name",
		OriginalRegion:    "us-east-1",
	}
	// mock out the region connection code
	state.Put("intermediary_image", false)
	stepAMIRegionCopy.getRegionConn = getMockConn
	stepAMIRegionCopy.Run(context.Background(), state)

	if len(stepAMIRegionCopy.Regions) != 0 {
		t.Fatalf("Should not have added original ami to Regions once; not" +
			"encrypting")
	}

	// ------------------------------------------------------------------------
	// Multiple regions, many duplicates, and encryption (this shouldn't ever
	// happen because of our template validation, but good to test it.)
	// ------------------------------------------------------------------------

	stepAMIRegionCopy = StepAMIRegionCopy{
		AccessConfig: testAccessConfig(),
		// Many duplicates for only 3 actual values
		Regions:     []string{"us-east-1", "us-west-2", "us-west-2", "ap-east-1", "ap-east-1", "ap-east-1"},
		AMIKmsKeyId: "IlikePancakes",
		// Original region key in regionkeyids is different than in amikmskeyid
		RegionKeyIds:      map[string]string{"us-east-1": "12345", "us-west-2": "abcde", "ap-east-1": "xyz"},
		EncryptBootVolume: config.TriTrue,
		Name:              "fake-ami-name",
		OriginalRegion:    "us-east-1",
	}
	// mock out the region connection code
	stepAMIRegionCopy.getRegionConn = getMockConn
	state.Put("intermediary_image", true)
	stepAMIRegionCopy.Run(context.Background(), state)

	if len(stepAMIRegionCopy.Regions) != 3 {
		t.Fatalf("Each AMI should have been added to Regions one time only.")
	}

	// Also verify that we respect RegionKeyIds over AMIKmsKeyIds:
	if stepAMIRegionCopy.RegionKeyIds["us-east-1"] != "12345" {
		t.Fatalf("RegionKeyIds should take precedence over AmiKmsKeyIds")
	}

	// ------------------------------------------------------------------------
	// Multiple regions, many duplicates, NO encryption
	// ------------------------------------------------------------------------

	stepAMIRegionCopy = StepAMIRegionCopy{
		AccessConfig: testAccessConfig(),
		// Many duplicates for only 3 actual values
		Regions:        []string{"us-east-1", "us-west-2", "us-west-2", "ap-east-1", "ap-east-1", "ap-east-1"},
		Name:           "fake-ami-name",
		OriginalRegion: "us-east-1",
	}
	// mock out the region connection code
	stepAMIRegionCopy.getRegionConn = getMockConn
	state.Put("intermediary_image", false)
	stepAMIRegionCopy.Run(context.Background(), state)

	if len(stepAMIRegionCopy.Regions) != 2 {
		t.Fatalf("Each AMI should have been added to Regions one time only, " +
			"and original region shouldn't be added at all")
	}
}

func TestStepAmiRegionCopy_nil_encryption(t *testing.T) {
	// create step
	stepAMIRegionCopy := StepAMIRegionCopy{
		AccessConfig:      testAccessConfig(),
		Regions:           make([]string, 0),
		AMIKmsKeyId:       "",
		RegionKeyIds:      make(map[string]string),
		EncryptBootVolume: config.TriUnset,
		Name:              "fake-ami-name",
		OriginalRegion:    "us-east-1",
	}
	// mock out the region connection code
	stepAMIRegionCopy.getRegionConn = getMockConn

	state := tState()
	state.Put("intermediary_image", false)
	stepAMIRegionCopy.Run(context.Background(), state)

	if stepAMIRegionCopy.toDelete != "" {
		t.Fatalf("Shouldn't have an intermediary ami if encrypt is nil")
	}
	if len(stepAMIRegionCopy.Regions) != 0 {
		t.Fatalf("Should not have added original ami to original region")
	}
}

func TestStepAmiRegionCopy_true_encryption(t *testing.T) {
	// create step
	stepAMIRegionCopy := StepAMIRegionCopy{
		AccessConfig:      testAccessConfig(),
		Regions:           make([]string, 0),
		AMIKmsKeyId:       "",
		RegionKeyIds:      make(map[string]string),
		EncryptBootVolume: config.TriTrue,
		Name:              "fake-ami-name",
		OriginalRegion:    "us-east-1",
	}
	// mock out the region connection code
	stepAMIRegionCopy.getRegionConn = getMockConn

	state := tState()
	state.Put("intermediary_image", true)
	stepAMIRegionCopy.Run(context.Background(), state)

	if stepAMIRegionCopy.toDelete == "" {
		t.Fatalf("Should delete original AMI if encrypted=true")
	}
	if len(stepAMIRegionCopy.Regions) == 0 {
		t.Fatalf("Should have added original ami to Regions")
	}
}

func TestStepAmiRegionCopy_nil_intermediary(t *testing.T) {
	// create step
	stepAMIRegionCopy := StepAMIRegionCopy{
		AccessConfig:      testAccessConfig(),
		Regions:           make([]string, 0),
		AMIKmsKeyId:       "",
		RegionKeyIds:      make(map[string]string),
		EncryptBootVolume: config.TriFalse,
		Name:              "fake-ami-name",
		OriginalRegion:    "us-east-1",
	}
	// mock out the region connection code
	stepAMIRegionCopy.getRegionConn = getMockConn

	state := tState()
	stepAMIRegionCopy.Run(context.Background(), state)

	if stepAMIRegionCopy.toDelete != "" {
		t.Fatalf("Should not delete original AMI if no intermediary")
	}
	if len(stepAMIRegionCopy.Regions) != 0 {
		t.Fatalf("Should not have added original ami to Regions")
	}
}

func TestStepAmiRegionCopy_AMISkipBuildRegion(t *testing.T) {
	// ------------------------------------------------------------------------
	// skip build region is true
	// ------------------------------------------------------------------------

	stepAMIRegionCopy := StepAMIRegionCopy{
		AccessConfig:       testAccessConfig(),
		Regions:            []string{"us-west-1"},
		AMIKmsKeyId:        "",
		RegionKeyIds:       map[string]string{"us-west-1": "abcde"},
		Name:               "fake-ami-name",
		OriginalRegion:     "us-east-1",
		AMISkipBuildRegion: true,
	}
	// mock out the region connection code
	stepAMIRegionCopy.getRegionConn = getMockConn

	state := tState()
	state.Put("intermediary_image", true)
	stepAMIRegionCopy.Run(context.Background(), state)

	if stepAMIRegionCopy.toDelete == "" {
		t.Fatalf("Should delete original AMI if skip_save_build_region=true")
	}
	if len(stepAMIRegionCopy.Regions) != 1 {
		t.Fatalf("Should not have added original ami to Regions; Regions: %#v", stepAMIRegionCopy.Regions)
	}

	// ------------------------------------------------------------------------
	// skip build region is false.
	// ------------------------------------------------------------------------
	stepAMIRegionCopy = StepAMIRegionCopy{
		AccessConfig:       testAccessConfig(),
		Regions:            []string{"us-west-1"},
		AMIKmsKeyId:        "",
		RegionKeyIds:       make(map[string]string),
		Name:               "fake-ami-name",
		OriginalRegion:     "us-east-1",
		AMISkipBuildRegion: false,
	}
	// mock out the region connection code
	stepAMIRegionCopy.getRegionConn = getMockConn

	state.Put("intermediary_image", false) // not encrypted
	stepAMIRegionCopy.Run(context.Background(), state)

	if stepAMIRegionCopy.toDelete != "" {
		t.Fatalf("Shouldn't have an intermediary AMI, so dont delete original ami")
	}
	if len(stepAMIRegionCopy.Regions) != 1 {
		t.Fatalf("Should not have added original ami to Regions; Regions: %#v", stepAMIRegionCopy.Regions)
	}

	// ------------------------------------------------------------------------
	// skip build region is false, but encrypt is true
	// ------------------------------------------------------------------------
	stepAMIRegionCopy = StepAMIRegionCopy{
		AccessConfig:       testAccessConfig(),
		Regions:            []string{"us-west-1"},
		AMIKmsKeyId:        "",
		RegionKeyIds:       map[string]string{"us-west-1": "abcde"},
		Name:               "fake-ami-name",
		OriginalRegion:     "us-east-1",
		AMISkipBuildRegion: false,
		EncryptBootVolume:  config.TriTrue,
	}
	// mock out the region connection code
	stepAMIRegionCopy.getRegionConn = getMockConn

	state.Put("intermediary_image", true) //encrypted
	stepAMIRegionCopy.Run(context.Background(), state)

	if stepAMIRegionCopy.toDelete == "" {
		t.Fatalf("Have to delete intermediary AMI")
	}
	if len(stepAMIRegionCopy.Regions) != 2 {
		t.Fatalf("Should have added original ami to Regions; Regions: %#v", stepAMIRegionCopy.Regions)
	}

	// ------------------------------------------------------------------------
	// skip build region is true, and encrypt is true
	// ------------------------------------------------------------------------
	stepAMIRegionCopy = StepAMIRegionCopy{
		AccessConfig:       testAccessConfig(),
		Regions:            []string{"us-west-1"},
		AMIKmsKeyId:        "",
		RegionKeyIds:       map[string]string{"us-west-1": "abcde"},
		Name:               "fake-ami-name",
		OriginalRegion:     "us-east-1",
		AMISkipBuildRegion: true,
		EncryptBootVolume:  config.TriTrue,
	}
	// mock out the region connection code
	stepAMIRegionCopy.getRegionConn = getMockConn

	state.Put("intermediary_image", true) //encrypted
	stepAMIRegionCopy.Run(context.Background(), state)

	if stepAMIRegionCopy.toDelete == "" {
		t.Fatalf("Have to delete intermediary AMI")
	}
	if len(stepAMIRegionCopy.Regions) != 1 {
		t.Fatalf("Should not have added original ami to Regions; Regions: %#v", stepAMIRegionCopy.Regions)
	}
}
