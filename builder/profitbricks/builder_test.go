package profitbricks

import (
	"testing"
	"github.com/mitchellh/packer/packer"
	"fmt"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"image": "Ubuntu-16.04",
		"pbpassword": "password",
		"pbusername": "username",
		"servername": "packer",
		"type": "profitbricks",
	}
}

func TestImplementsBuilder (t *testing.T){
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

	warns, err := b.Prepare(c)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		fmt.Println(err)
		fmt.Println(warns)
		t.Fatalf("prepare should fail")
	}
}

func TestBuilderPrepare_InvalidKey(t *testing.T) {
	var b Builder
	config := testConfig()

	config["i_should_not_be_valid"] = true
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_Servername(t *testing.T) {
	var b Builder
	config := testConfig()

	delete(config, "servername")
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatalf("should error")
	}

	expected := "packer"

	config["servername"] = expected
	b = Builder{}
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if b.config.ServerName != expected {
		t.Errorf("found %s, expected %s", b.config.ServerName, expected)
	}
}