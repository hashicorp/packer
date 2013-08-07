package fix

import (
	"reflect"
	"testing"
)

func TestFixerGlobalTemplates_Impl(t *testing.T) {
	var raw interface{}
	raw = new(FixerGlobalTemplates)
	if _, ok := raw.(Fixer); !ok {
		t.Fatalf("must be a Fixer")
	}
}

func TestFixerGlobalTemplatesFix_DigitalOcean(t *testing.T) {
	var f FixerGlobalTemplates

	input := map[string]interface{}{
		"builders": []interface{}{
			map[string]string{
				"type":          "digitalocean",
				"snapshot_name": "foo-{{.CreateTime}}",
			},
		},
	}

	expected := map[string]interface{}{
		"builders": []map[string]interface{}{
			map[string]interface{}{
				"type":          "digitalocean",
				"snapshot_name": "foo-{{timestamp}}",
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

func TestFixerGlobalTemplatesFix_VirtualBox(t *testing.T) {
	var f FixerGlobalTemplates

	input := map[string]interface{}{
		"builders": []interface{}{
			map[string]string{
				"type":       "virtualbox",
				"output_dir": "foo-{{.CreateTime}}",
				"vm_name":    "foo-{{.HTTPIP}}",
			},
		},
	}

	expected := map[string]interface{}{
		"builders": []map[string]interface{}{
			map[string]interface{}{
				"type":       "virtualbox",
				"output_dir": "foo-{{timestamp}}",
				"vm_name":    `foo-{{builder "http_ip"}}`,
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

func TestFixerGlobalTemplatesFix_VMware(t *testing.T) {
	var f FixerGlobalTemplates

	input := map[string]interface{}{
		"builders": []interface{}{
			map[string]string{
				"type":       "virtualbox",
				"output_dir": "foo-{{.CreateTime}}",
				"vm_name":    "foo-{{.HTTPIP}}",
			},
		},
	}

	expected := map[string]interface{}{
		"builders": []map[string]interface{}{
			map[string]interface{}{
				"type":       "virtualbox",
				"output_dir": "foo-{{timestamp}}",
				"vm_name":    `foo-{{builder "http_ip"}}`,
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
