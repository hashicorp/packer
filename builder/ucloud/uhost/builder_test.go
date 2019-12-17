package uhost

import (
	"reflect"
	"testing"

	ucloudcommon "github.com/hashicorp/packer/builder/ucloud/common"
	"github.com/hashicorp/packer/packer"
)

func testBuilderConfig() map[string]interface{} {
	return map[string]interface{}{
		"public_key":        "foo",
		"private_key":       "bar",
		"project_id":        "foo",
		"source_image_id":   "bar",
		"availability_zone": "cn-bj2-02",
		"instance_type":     "n-basic-2",
		"region":            "cn-bj2",
		"ssh_username":      "root",
		"image_name":        "foo",
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
		"public_key": []string{},
	}

	warnings, err := b.Prepare(c)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatalf("prepare should fail")
	}
}

func TestBuilderPrepare_ImageName(t *testing.T) {
	var b Builder
	config := testBuilderConfig()

	// Test good
	config["image_name"] = "foo"
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test bad
	config["image_name"] = "foo {{"
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

func TestBuilderPrepare_ImageDestinations(t *testing.T) {
	var b Builder
	config := testBuilderConfig()
	config["image_copy_to_mappings"] = []map[string]interface{}{
		{
			"project_id":  "project1",
			"region":      "region1",
			"name":        "bar",
			"description": "foo",
		},
		{
			"project_id":  "project2",
			"region":      "region2",
			"name":        "foo",
			"description": "bar",
		},
	}
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if !reflect.DeepEqual(b.config.ImageDestinations, []ucloudcommon.ImageDestination{
		{
			ProjectId:   "project1",
			Region:      "region1",
			Name:        "bar",
			Description: "foo",
		},
		{
			ProjectId:   "project2",
			Region:      "region2",
			Name:        "foo",
			Description: "bar",
		},
	}) {
		t.Fatalf("image_copy_mappings are not set properly, got: %#v", b.config.ImageDestinations)
	}
}
