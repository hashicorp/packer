package ebsvolume

import (
	"bytes"
	"context"
	"sync"
	"testing"

	"github.com/aws/aws-sdk-go/aws"

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

	// Counters to figure out what code path was taken
	copyImageCount       int
	describeImagesCount  int
	deregisterImageCount int
	deleteSnapshotCount  int
	waitCount            int

	lock sync.Mutex
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
	state.Put("instance", &ec2.Instance{
		InstanceId: aws.String("instance-id"),
	})

	accessConfig := common.FakeAccessConfig()

	volMap := BlockDevices{
		{
	awscommon.BlockDevice `mapstructure:",squash"`
	// Key/value pair tags to apply to the volume. These are retained after the builder
	// completes. This is a [template engine](/docs/templates/legacy_json_templates/engine), see
	// [Build template data](#build-template-data) for more information.
	Tags map[string]string `mapstructure:"tags" required:"false"`
	// Same as [`tags`](#tags) but defined as a singular repeatable block
	// containing a `key` and a `value` field. In HCL2 mode the
	// [`dynamic_block`](/docs/templates/hcl_templates/expressions#dynamic-blocks)
	// will allow you to create those programatically.
	Tag config.KeyValues `mapstructure:"tag" required:"false"`

	// Create a Snapshot of this Volume.
	SnapshotVolume bool `mapstructure:"snapshot_volume" required:"false"`

	awscommon.SnapshotConfig `mapstructure:",squash"`
}
}
	//Todo add fake volumes, for the snap shot step to Snapshot

	step := stepSnapshotEBSVolumes{
		PollingConfig: new(common.AWSPollingConfig), //Dosnt look like builder sets this up
		AccessConfig:  accessConfig,
		VolumeMapping: b.config.VolumeMappings,
		Ctx:           b.config.ctx,
	}

	step.Run(context.Background(), state)

	if len(step.snapshotMap) != 1 {
		t.Fatalf("Missing Snapshot from step")
	}
}
