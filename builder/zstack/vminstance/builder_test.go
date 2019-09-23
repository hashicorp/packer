package vminstance

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func BuilderConfigTest() map[string]interface{} {
	return map[string]interface{}{
		"access_key":        "foo",
		"key_secret":        "bar",
		"image_uuid":        "image",
		"l3network_uuid":    "l3",
		"instance_offering": "instance",
		"zone_uuid":         "zone",
		"ssh_username":      "root",
		"base_url":          "http://192.168.0.1:8080",
	}
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packer.Builder); !ok {
		t.Fatalf("Builder should be a builder")
	}
}

var reqired_list = []string{"image_uuid", "l3network_uuid", "instance_offering", "base_url", "access_key",
	"key_secret", "zone_uuid", "ssh_username"}

func TestBuilder_Prepare_BadType(t *testing.T) {
	b := &Builder{}
	c := map[string]interface{}{
		"access_key": "string",
	}

	warnings, err := b.Prepare(c)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatalf("prepare should fail")
	}

	flag := false
	for _, reqired := range reqired_list {
		if !strings.Contains(err.Error(), reqired) && reqired != "access_key" {
			fmt.Printf("err should include '%s'", reqired)
			flag = true
		}
	}
	if flag {
		t.Fatal()
	}
}

func TestBuilder_Prepare_CorrectType(t *testing.T) {
	b := &Builder{}
	c := map[string]interface{}{}
	for _, reqired := range reqired_list {
		c[reqired] = "123"
	}
	_, err := b.Prepare(c)
	if err != nil {
		t.Fatalf("prepare should success")
	}
}

func TestBuilderPrepare_InvalidKey(t *testing.T) {
	var b Builder
	config := BuilderConfigTest()

	// Add a random key
	config["i_should_not_be_valid"] = true
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil || !strings.Contains(err.Error(), "unknown configuration key") {
		t.Fatal("should have error")
	}
}
