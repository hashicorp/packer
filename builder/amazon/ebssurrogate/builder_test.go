package ebssurrogate

import (
	"testing"

	"github.com/hashicorp/packer/builder/amazon/common"

	"github.com/hashicorp/packer/packer"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"access_key":    "foo",
		"secret_key":    "bar",
		"source_ami":    "foo",
		"instance_type": "foo",
		"region":        "us-east-1",
		"ssh_username":  "root",
	}
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packer.Builder); !ok {
		t.Fatal("Builder should be a builder")
	}
}

func TestBuilder_Prepare_BadType(t *testing.T) {
	b := &Builder{}
	c := map[string]interface{}{
		"access_key": []string{},
	}

	_, warnings, err := b.Prepare(c)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("prepare should fail")
	}
}

func TestBuilderPrepare_InvalidKey(t *testing.T) {
	var b Builder
	config := testConfig()

	// Add a random key
	config["i_should_not_be_valid"] = true
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_ReturnGeneratedData(t *testing.T) {
	var b Builder
	// Basic configuration
	b.config.RootDevice = RootBlockDevice{
		SourceDeviceName: "device name",
		DeviceName:       "device name",
	}
	b.config.LaunchMappings = BlockDevices{
		BlockDevice{
			BlockDevice: common.BlockDevice{
				DeviceName: "device name",
			},
			OmitFromArtifact: false,
		},
	}
	b.config.AMIVirtType = "type"
	config := testConfig()
	config["ami_name"] = "name"

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
	if generatedData[0] != "SourceAMIName" {
		t.Fatalf("Generated data should contain SourceAMIName")
	}
}
