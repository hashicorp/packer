package fix

import (
	"reflect"
	"testing"
)

func TestFixerVagrantPPOverride_Impl(t *testing.T) {
	var _ Fixer = new(FixerVagrantPPOverride)
}

func TestFixerVagrantPPOverride_Fix(t *testing.T) {
	var f FixerVagrantPPOverride

	input := map[string]interface{}{
		"post-processors": []interface{}{
			"foo",

			map[string]interface{}{
				"type": "vagrant",
				"aws": map[string]interface{}{
					"foo": "bar",
				},
			},

			map[string]interface{}{
				"type": "vsphere",
			},

			[]interface{}{
				map[string]interface{}{
					"type": "vagrant",
					"vmware": map[string]interface{}{
						"foo": "bar",
					},
				},
			},
		},
	}

	expected := map[string]interface{}{
		"post-processors": []interface{}{
			"foo",

			map[string]interface{}{
				"type": "vagrant",
				"override": map[string]interface{}{
					"aws": map[string]interface{}{
						"foo": "bar",
					},
				},
			},

			map[string]interface{}{
				"type": "vsphere",
			},

			[]interface{}{
				map[string]interface{}{
					"type": "vagrant",
					"override": map[string]interface{}{
						"vmware": map[string]interface{}{
							"foo": "bar",
						},
					},
				},
			},
		},
	}

	output, err := f.Fix(input)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !reflect.DeepEqual(output, expected) {
		t.Fatalf("unexpected: %#v\nexpected: %#v\n", output, expected)
	}
}
