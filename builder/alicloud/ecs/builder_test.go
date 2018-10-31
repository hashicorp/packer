package ecs

import (
	"reflect"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func testBuilderConfig() map[string]interface{} {
	return map[string]interface{}{
		"access_key":    "foo",
		"secret_key":    "bar",
		"source_image":  "foo",
		"instance_type": "ecs.n1.tiny",
		"region":        "cn-beijing",
		"ssh_username":  "root",
		"image_name":    "foo",
		"io_optimized":  true,
	}
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packer.Builder); !ok {
		t.Fatalf("Builder should be a builder")
	}
}

func TestBuilder_Prepare_BadType(t *testing.T) {
	b := &Builder{}
	c := map[string]interface{}{
		"access_key": []string{},
	}

	warnings, err := b.Prepare(c)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatalf("prepare should fail")
	}
}

func TestBuilderPrepare_ECSImageName(t *testing.T) {
	var b Builder
	config := testBuilderConfig()

	// Test good
	config["image_name"] = "ecs.n1.tiny"
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test bad
	config["ecs_image_name"] = "foo {{"
	b = Builder{}
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test bad
	delete(config, "image_name")
	b = Builder{}
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_InvalidKey(t *testing.T) {
	var b Builder
	config := testBuilderConfig()

	// Add a random key
	config["i_should_not_be_valid"] = true
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_Devices(t *testing.T) {
	var b Builder
	config := testBuilderConfig()
	config["system_disk_mapping"] = map[string]interface{}{
		"disk_category":    "cloud",
		"disk_description": "system disk",
		"disk_name":        "system_disk",
		"disk_size":        60,
	}
	config["image_disk_mappings"] = []map[string]interface{}{
		{
			"disk_category":             "cloud_efficiency",
			"disk_name":                 "data_disk1",
			"disk_size":                 100,
			"disk_snapshot_id":          "s-1",
			"disk_description":          "data disk1",
			"disk_device":               "/dev/xvdb",
			"disk_delete_with_instance": false,
		},
		{
			"disk_name":   "data_disk2",
			"disk_device": "/dev/xvdc",
		},
	}
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	if !reflect.DeepEqual(b.config.ECSSystemDiskMapping, AlicloudDiskDevice{
		DiskCategory: "cloud",
		Description:  "system disk",
		DiskName:     "system_disk",
		DiskSize:     60,
	}) {
		t.Fatalf("system disk is not set properly, actual: %#v", b.config.ECSSystemDiskMapping)
	}
	if !reflect.DeepEqual(b.config.ECSImagesDiskMappings, []AlicloudDiskDevice{
		{
			DiskCategory:       "cloud_efficiency",
			DiskName:           "data_disk1",
			DiskSize:           100,
			SnapshotId:         "s-1",
			Description:        "data disk1",
			Device:             "/dev/xvdb",
			DeleteWithInstance: false,
		},
		{
			DiskName: "data_disk2",
			Device:   "/dev/xvdc",
		},
	}) {
		t.Fatalf("data disks are not set properly, actual: %#v", b.config.ECSImagesDiskMappings)
	}
}
