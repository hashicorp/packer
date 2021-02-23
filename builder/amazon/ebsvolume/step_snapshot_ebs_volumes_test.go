package ebsvolume

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"

	//"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/amazon/common"
)

// Define a mock struct to be used in unit tests for common aws steps.
type mockEC2Conn struct {
	ec2iface.EC2API
	Config *aws.Config
}

func (m *mockEC2Conn) CreateSnapshot(input *ec2.CreateSnapshotInput) (*ec2.Snapshot, error) {
	snap := &ec2.Snapshot{
		// This isn't typical amazon format, but injecting the volume id into
		// this field lets us verify that the right volume was snapshotted with
		// a simple string comparison
		SnapshotId: aws.String(fmt.Sprintf("snap-of-%s", *input.VolumeId)),
	}

	return snap, nil
}

func (m *mockEC2Conn) WaitUntilSnapshotCompletedWithContext(aws.Context, *ec2.DescribeSnapshotsInput, ...request.WaiterOption) error {
	return nil
}

func getMockConn(config *common.AccessConfig, target string) (ec2iface.EC2API, error) {
	mockConn := &mockEC2Conn{
		Config: aws.NewConfig(),
	}
	return mockConn, nil
}

// Create statebag for running test
func tState(t *testing.T) multistep.StateBag {
	state := new(multistep.BasicStateBag)
	state.Put("ui", &packer.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	// state.Put("amis", map[string]string{"us-east-1": "ami-12345"})
	// state.Put("snapshots", map[string][]string{"us-east-1": {"snap-0012345"}})
	conn, _ := getMockConn(&common.AccessConfig{}, "us-east-2")

	state.Put("ec2", conn)
	// Store a fake instance that contains a block device that matches the
	// volumes defined in the config above
	state.Put("instance", &ec2.Instance{
		InstanceId: aws.String("instance-id"),
		BlockDeviceMappings: []*ec2.InstanceBlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/xvda"),
				Ebs: &ec2.EbsInstanceBlockDevice{
					VolumeId: aws.String("vol-1234"),
				},
			},
			{
				DeviceName: aws.String("/dev/xvdb"),
				Ebs: &ec2.EbsInstanceBlockDevice{
					VolumeId: aws.String("vol-5678"),
				},
			},
		},
	})
	return state
}

func TestStepSnapshot_run_simple(t *testing.T) {
	var b Builder
	config := testConfig() //from builder_test

	//Set some snapshot settings
	config["ebs_volumes"] = []map[string]interface{}{
		{
			"device_name":           "/dev/xvdb",
			"volume_size":           "32",
			"delete_on_termination": true,
			"snapshot_volume":       true,
		},
	}

	generatedData, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	if len(generatedData) == 0 {
		t.Fatalf("Generated data should not be empty")
	}

	state := tState(t)

	accessConfig := common.FakeAccessConfig()

	step := stepSnapshotEBSVolumes{
		PollingConfig: new(common.AWSPollingConfig),
		AccessConfig:  accessConfig,
		VolumeMapping: b.config.VolumeMappings,
		Ctx:           b.config.ctx,
	}

	step.Run(context.Background(), state)

	if len(step.snapshotMap) != 1 {
		t.Fatalf("Missing Snapshot from step")
	}

	if volmapping := step.snapshotMap["snap-of-vol-5678"]; volmapping == nil {
		t.Fatalf("Didn't snapshot correct volume: Map is %#v", step.snapshotMap)
	}
}

func TestStepSnapshot_run_no_snaps(t *testing.T) {
	var b Builder
	config := testConfig() //from builder_test

	//Set some snapshot settings
	config["ebs_volumes"] = []map[string]interface{}{
		{
			"device_name":           "/dev/xvdb",
			"volume_size":           "32",
			"delete_on_termination": true,
			"snapshot_volume":       false,
		},
	}

	generatedData, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	if len(generatedData) == 0 {
		t.Fatalf("Generated data should not be empty")
	}

	state := tState(t)

	accessConfig := common.FakeAccessConfig()

	step := stepSnapshotEBSVolumes{
		PollingConfig: new(common.AWSPollingConfig),
		AccessConfig:  accessConfig,
		VolumeMapping: b.config.VolumeMappings,
		Ctx:           b.config.ctx,
	}

	step.Run(context.Background(), state)

	if len(step.snapshotMap) != 0 {
		t.Fatalf("Shouldn't have snapshotted any volumes")
	}
}
