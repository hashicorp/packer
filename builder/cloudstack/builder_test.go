package cloudstack

import (
	"testing"

	"github.com/mitchellh/packer/packer"
)

func TestBuilder_Impl(t *testing.T) {
	var raw interface{} = &Builder{}

	if _, ok := raw.(packer.Builder); !ok {
		t.Fatalf("Builder does not implement packer.Builder")
	}
}

func TestBuilder_Prepare(t *testing.T) {
	cases := map[string]struct {
		Config map[string]interface{}
		Err    bool
	}{
		"good": {
			Config: map[string]interface{}{
				"api_url":          "https://cloudstack.com/client/api",
				"api_key":          "some-api-key",
				"secret_key":       "some-secret-key",
				"cidr_list":        []interface{}{"0.0.0.0/0"},
				"disk_size":        "20",
				"network":          "c5ed8a14-3f21-4fa9-bd74-bb887fc0ed0d",
				"service_offering": "a29c52b1-a83d-4123-a57d-4548befa47a0",
				"source_template":  "d31e6af5-94a8-4756-abf3-6493c38db7e5",
				"ssh_username":     "ubuntu",
				"template_os":      "52d54d24-cef1-480b-b963-527703aa4ff9",
				"zone":             "a3b594d9-25e9-47c1-9c03-7a5fc61e3f43",
			},
			Err: false,
		},
		"bad": {
			Err: true,
		},
	}

	b := &Builder{}

	for desc, tc := range cases {
		_, errs := b.Prepare(tc.Config)

		if tc.Err {
			if errs == nil {
				t.Fatalf("%s prepare should err", desc)
			}
		} else {
			if errs != nil {
				t.Fatalf("%s prepare should not fail: %s", desc, errs)
			}
		}
	}
}
