package openstack

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"username":     "foo",
		"password":     "bar",
		"region":       "DFW",
		"image_name":   "foo",
		"source_image": "foo",
		"flavor":       "foo",
		"ssh_username": "root",
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
		"password": []string{},
	}

	warns, err := b.Prepare(c)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatalf("prepare should fail")
	}
}
