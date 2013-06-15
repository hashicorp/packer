package digitalocean

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"client_id": "foo",
		"api_key":   "bar",
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
		"api_key": []string{},
	}

	err := b.Prepare(c)
	if err == nil {
		t.Fatalf("prepare should fail")
	}
}

func TestBuilderPrepare_APIKey(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test good
	config["api_key"] = "foo"
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.APIKey != "foo" {
		t.Errorf("access key invalid: %s", b.config.APIKey)
	}

	// Test bad
	delete(config, "api_key")
	b = Builder{}
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_ClientID(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test good
	config["client_id"] = "foo"
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.ClientID != "foo" {
		t.Errorf("invalid: %s", b.config.ClientID)
	}

	// Test bad
	delete(config, "client_id")
	b = Builder{}
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_RegionID(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.RegionID != 1 {
		t.Errorf("invalid: %d", b.config.RegionID)
	}

	// Test set
	config["region_id"] = 2
	b = Builder{}
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.RegionID != 2 {
		t.Errorf("invalid: %d", b.config.RegionID)
	}
}

func TestBuilderPrepare_SizeID(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.SizeID != 66 {
		t.Errorf("invalid: %d", b.config.SizeID)
	}

	// Test set
	config["size_id"] = 67
	b = Builder{}
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.SizeID != 67 {
		t.Errorf("invalid: %d", b.config.SizeID)
	}
}

func TestBuilderPrepare_ImageID(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.SizeID != 2676 {
		t.Errorf("invalid: %d", b.config.SizeID)
	}

	// Test set
	config["size_id"] = 2
	b = Builder{}
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.SizeID != 2 {
		t.Errorf("invalid: %d", b.config.SizeID)
	}
}

func TestBuilderPrepare_SSHUsername(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.SSHUsername != "root" {
		t.Errorf("invalid: %d", b.config.SSHUsername)
	}

	// Test set
	config["ssh_username"] = ""
	b = Builder{}
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.SSHPort != 35 {
		t.Errorf("invalid: %d", b.config.SSHPort)
	}
}

func TestBuilderPrepare_SSHTimeout(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.RawSSHTimeout != "1m" {
		t.Errorf("invalid: %d", b.config.RawSSHTimeout)
	}

	// Test set
	config["ssh_timeout"] = "30s"
	b = Builder{}
	err = b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test bad
	config["ssh_timeout"] = "tubes"
	b = Builder{}
	err = b.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

}

func TestBuilderPrepare_SnapshotName(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test set
	config["snapshot_name"] = "foo"
	err := b.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.SnapshotName != "foo" {
		t.Errorf("invalid: %s", b.config.SnapshotName)
	}
}
