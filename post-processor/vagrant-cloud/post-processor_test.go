package vagrantcloud

import (
	"bytes"
	"github.com/mitchellh/packer/packer"
	"testing"
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

func TestproviderFromBuilderName(t *testing.T) {
	if providerFromBuilderName("foobar") != "foobar" {
		t.Fatal("should copy unknown provider")
	}

	if providerFromBuilderName("vmware") != "vmware_desktop" {
		t.Fatal("should convert provider")
	}
}
