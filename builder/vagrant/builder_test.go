package vagrant

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packer.Builder); !ok {
		t.Fatalf("Builder should be a builder")
	}
}

func TestBuilder_Prepare_ValidateSource(t *testing.T) {
	b := &Builder{}
	type testCase struct {
		config      map[string]interface{}
		errExpected bool
		reason      string
	}

	cases := []testCase{
		{
			config: map[string]interface{}{
				"global_id": "a3559ec",
			},
			errExpected: true,
			reason:      "Need to set SSH communicator.",
		},
		{
			config: map[string]interface{}{
				"global_id":    "a3559ec",
				"communicator": "ssh",
			},
			errExpected: false,
			reason:      "Shouldn't fail because we've set global_id",
		},
		{
			config: map[string]interface{}{
				"communicator": "ssh",
			},
			errExpected: true,
			reason:      "Should fail because we must set source_path or global_id",
		},
		{
			config: map[string]interface{}{
				"source_path":  "./mybox",
				"communicator": "ssh",
			},
			errExpected: false,
			reason:      "Source path is set; we should be fine",
		},
		{
			config: map[string]interface{}{
				"source_path":  "./mybox",
				"communicator": "ssh",
				"global_id":    "a3559ec",
			},
			errExpected: true,
			reason:      "Both source path and global are set: we should error.",
		},
		{
			config: map[string]interface{}{
				"communicator":    "ssh",
				"global_id":       "a3559ec",
				"teardown_method": "suspend",
			},
			errExpected: false,
			reason:      "Valid argument for teardown method",
		},
		{
			config: map[string]interface{}{
				"communicator":    "ssh",
				"global_id":       "a3559ec",
				"teardown_method": "surspernd",
			},
			errExpected: true,
			reason:      "Inalid argument for teardown method",
		},
	}

	for _, tc := range cases {
		_, _, err := b.Prepare(tc.config)
		if (err != nil) != tc.errExpected {
			t.Fatalf("Unexpected behavior from test case %#v; %s.", tc.config, tc.reason)
		}
	}
}
