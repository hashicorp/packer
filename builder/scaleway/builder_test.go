package scaleway

import (
	"strconv"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"project_id":      "00000000-1111-2222-3333-444444444444",
		"access_key":      "SCWABCXXXXXXXXXXXXXX",
		"secret_key":      "00000000-1111-2222-3333-444444444444",
		"zone":            "fr-par-1",
		"commercial_type": "START1-S",
		"ssh_username":    "root",
		"image":           "image-uuid",
	}
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packersdk.Builder); !ok {
		t.Fatalf("Builder should be a builder")
	}
}

func TestBuilder_Prepare_BadType(t *testing.T) {
	b := &Builder{}
	c := map[string]interface{}{
		"api_token": []string{},
	}

	_, _, err := b.Prepare(c)
	if err == nil {
		t.Fatalf("prepare should fail")
	}
}

func TestBuilderPrepare(t *testing.T) {
	var b Builder
	config := testConfig()

	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatal("should not have errors")
	}
}

func TestBuilderPrepare_InvalidKey(t *testing.T) {
	var b Builder
	config := testConfig()

	config["i_should_not_be_valid"] = true
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_Zone(t *testing.T) {
	var b Builder
	config := testConfig()

	delete(config, "zone")
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatalf("should error")
	}

	expected := "fr-par-1"

	config["zone"] = expected
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.Zone != expected {
		t.Errorf("found %s, expected %s", b.config.Zone, expected)
	}
}

func TestBuilderPrepare_CommercialType(t *testing.T) {
	var b Builder
	config := testConfig()

	delete(config, "commercial_type")
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatalf("should error")
	}

	expected := "START1-S"

	config["commercial_type"] = expected
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.CommercialType != expected {
		t.Errorf("found %s, expected %s", b.config.CommercialType, expected)
	}
}

func TestBuilderPrepare_Image(t *testing.T) {
	var b Builder
	config := testConfig()

	delete(config, "image")
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should error")
	}

	expected := "cc586e45-5156-4f71-b223-cf406b10dd1c"

	config["image"] = expected
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.Image != expected {
		t.Errorf("found %s, expected %s", b.config.Image, expected)
	}
}

func TestBuilderPrepare_SnapshotName(t *testing.T) {
	var b Builder
	config := testConfig()

	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.SnapshotName == "" {
		t.Errorf("invalid: %s", b.config.SnapshotName)
	}

	config["snapshot_name"] = "foobarbaz"
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	config["snapshot_name"] = "{{timestamp}}"
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	_, err = strconv.ParseInt(b.config.SnapshotName, 0, 0)
	if err != nil {
		t.Fatalf("failed to parse int in template: %s", err)
	}

}

func TestBuilderPrepare_ServerName(t *testing.T) {
	var b Builder
	config := testConfig()

	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.ServerName == "" {
		t.Errorf("invalid: %s", b.config.ServerName)
	}

	config["server_name"] = "foobar"
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	config["server_name"] = "foobar-{{timestamp}}"
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	config["server_name"] = "foobar-{{"
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}

}
