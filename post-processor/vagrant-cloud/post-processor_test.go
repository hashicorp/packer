package vagrantcloud

import (
	"bytes"
	"os"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func testGoodConfig() map[string]interface{} {
	return map[string]interface{}{
		"access_token":        "foo",
		"version_description": "bar",
		"box_tag":             "hashicorp/precise64",
		"version":             "0.5",
	}
}

func testBadConfig() map[string]interface{} {
	return map[string]interface{}{
		"access_token":        "foo",
		"box_tag":             "baz",
		"version_description": "bar",
	}
}

func TestPostProcessor_Configure_fromVagrantEnv(t *testing.T) {
	var p PostProcessor
	config := testGoodConfig()
	config["access_token"] = ""
	os.Setenv("VAGRANT_CLOUD_TOKEN", "bar")
	defer func() {
		os.Setenv("VAGRANT_CLOUD_TOKEN", "")
	}()

	if err := p.Configure(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.AccessToken != "bar" {
		t.Fatalf("Expected to get token from VAGRANT_CLOUD_TOKEN env var. Got '%s' instead",
			p.config.AccessToken)
	}
}

func TestPostProcessor_Configure_fromAtlasEnv(t *testing.T) {
	var p PostProcessor
	config := testGoodConfig()
	config["access_token"] = ""
	os.Setenv("ATLAS_TOKEN", "foo")
	defer func() {
		os.Setenv("ATLAS_TOKEN", "")
	}()

	if err := p.Configure(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.AccessToken != "foo" {
		t.Fatalf("Expected to get token from ATLAS_TOKEN env var. Got '%s' instead",
			p.config.AccessToken)
	}

	if !p.warnAtlasToken {
		t.Fatal("Expected warn flag to be set when getting token from atlas env var.")
	}
}

func TestPostProcessor_Configure_Good(t *testing.T) {
	var p PostProcessor
	if err := p.Configure(testGoodConfig()); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestPostProcessor_Configure_Bad(t *testing.T) {
	var p PostProcessor
	if err := p.Configure(testBadConfig()); err == nil {
		t.Fatalf("should have err")
	}
}

func testUi() *packer.BasicUi {
	return &packer.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	}
}

func TestPostProcessor_ImplementsPostProcessor(t *testing.T) {
	var _ packer.PostProcessor = new(PostProcessor)
}

func TestProviderFromBuilderName(t *testing.T) {
	if providerFromBuilderName("foobar") != "foobar" {
		t.Fatal("should copy unknown provider")
	}

	if providerFromBuilderName("vmware") != "vmware_desktop" {
		t.Fatal("should convert provider")
	}
}
