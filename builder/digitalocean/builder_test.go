package digitalocean

import (
	"os"
	"strconv"
	"testing"

	"github.com/mitchellh/packer/packer"
)

func init() {
	// Clear out the credential env vars
	os.Setenv("DIGITALOCEAN_API_KEY", "")
	os.Setenv("DIGITALOCEAN_CLIENT_ID", "")
	os.Setenv("DIGITALOCEAN_API_TOKEN", "")
}

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"client_id": "foo",
		"api_key":   "bar",
		"api_url":   "http://example.com",
		"api_token": "token",
		"image":     "ubuntu-14-04-x64",
		"size":      "512m",
		"region":    "nyc3",
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

	warnings, err := b.Prepare(c)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatalf("prepare should fail")
	}
}

func TestBuilderPrepare_APIToken(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test good
	config["api_token"] = "token"
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.APIToken != "token" {
		t.Errorf("access token invalid: %s", b.config.APIToken)
	}

	// Test bad
	delete(config, "api_token")
	delete(config, "api_key")
	b = Builder{}
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test env variable
	delete(config, "api_token")
	os.Setenv("DIGITALOCEAN_API_TOKEN", "token")
	defer os.Setenv("DIGITALOCEAN_API_TOKEN", "")
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_APIKey(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test good
	config["api_key"] = "foo"
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.APIKey != "foo" {
		t.Errorf("access key invalid: %s", b.config.APIKey)
	}

	// Test bad
	delete(config, "api_token")
	delete(config, "api_key")
	b = Builder{}
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test env variable
	delete(config, "api_key")
	os.Setenv("DIGITALOCEAN_API_KEY", "foo")
	defer os.Setenv("DIGITALOCEAN_API_KEY", "")
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_ClientID(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test good
	config["client_id"] = "foo"
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.ClientID != "foo" {
		t.Errorf("invalid: %s", b.config.ClientID)
	}

	// Test bad
	delete(config, "api_token")
	delete(config, "client_id")
	b = Builder{}
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test env variable
	delete(config, "client_id")
	os.Setenv("DIGITALOCEAN_CLIENT_ID", "foo")
	defer os.Setenv("DIGITALOCEAN_CLIENT_ID", "")
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_InvalidKey(t *testing.T) {
	var b Builder
	config := testConfig()

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

func TestBuilderPrepare_Region(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	expected := "sfo1"

	// Test set
	config["region"] = expected
	b = Builder{}
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.Region != expected {
		t.Errorf("found %s, expected %s", b.config.Region, expected)
	}
}

func TestBuilderPrepare_Size(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	expected := "1024mb"

	// Test set
	config["size"] = expected
	b = Builder{}
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.Size != expected {
		t.Errorf("found %s, expected %s", b.config.Size, expected)
	}
}

func TestBuilderPrepare_Image(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	expected := "ubuntu-14-04-x64"

	// Test set
	config["image"] = expected
	b = Builder{}
	warnings, err = b.Prepare(config)
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

func TestBuilderPrepare_SSHUsername(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.SSHUsername != "root" {
		t.Errorf("invalid: %s", b.config.SSHUsername)
	}

	// Test set
	config["ssh_username"] = "foo"
	b = Builder{}
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.SSHUsername != "foo" {
		t.Errorf("invalid: %s", b.config.SSHUsername)
	}
}

func TestBuilderPrepare_SSHTimeout(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.RawSSHTimeout != "1m" {
		t.Errorf("invalid: %s", b.config.RawSSHTimeout)
	}

	// Test set
	config["ssh_timeout"] = "30s"
	b = Builder{}
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test bad
	config["ssh_timeout"] = "tubes"
	b = Builder{}
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}

}

func TestBuilderPrepare_StateTimeout(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.RawStateTimeout != "6m" {
		t.Errorf("invalid: %s", b.config.RawStateTimeout)
	}

	// Test set
	config["state_timeout"] = "5m"
	b = Builder{}
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test bad
	config["state_timeout"] = "tubes"
	b = Builder{}
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}

}

func TestBuilderPrepare_PrivateNetworking(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.PrivateNetworking != false {
		t.Errorf("invalid: %t", b.config.PrivateNetworking)
	}

	// Test set
	config["private_networking"] = true
	b = Builder{}
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.PrivateNetworking != true {
		t.Errorf("invalid: %t", b.config.PrivateNetworking)
	}
}

func TestBuilderPrepare_SnapshotName(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.SnapshotName == "" {
		t.Errorf("invalid: %s", b.config.SnapshotName)
	}

	// Test set
	config["snapshot_name"] = "foobarbaz"
	b = Builder{}
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test set with template
	config["snapshot_name"] = "{{timestamp}}"
	b = Builder{}
	warnings, err = b.Prepare(config)
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

func TestBuilderPrepare_DropletName(t *testing.T) {
	var b Builder
	config := testConfig()

	// Test default
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.DropletName == "" {
		t.Errorf("invalid: %s", b.config.DropletName)
	}

	// Test normal set
	config["droplet_name"] = "foobar"
	b = Builder{}
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test with template
	config["droplet_name"] = "foobar-{{timestamp}}"
	b = Builder{}
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test with bad template
	config["droplet_name"] = "foobar-{{"
	b = Builder{}
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}

}
