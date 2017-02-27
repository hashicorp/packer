package fix

import (
	"reflect"
	"testing"
)

func TestFixerManifestPPFilename_Impl(t *testing.T) {
	var _ Fixer = new(FixerVagrantPPOverride)
}

func TestFixerManifestPPFilename_Fix(t *testing.T) {
	var f FixerManifestFilename

	input := map[string]interface{}{
		"post-processors": []interface{}{
			map[string]interface{}{
				"type":     "manifest",
				"filename": "foo",
			},
			[]interface{}{
				map[string]interface{}{
					"type":     "manifest",
					"filename": "foo",
				},
			},
		},
	}

	expected := map[string]interface{}{
		"post-processors": []interface{}{
			map[string]interface{}{
				"type":   "manifest",
				"output": "foo",
			},
			[]interface{}{
				map[string]interface{}{
					"type":   "manifest",
					"output": "foo",
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
