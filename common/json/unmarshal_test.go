package json

import "testing"

func TestUnmarshalJson(t *testing.T) {
	b := []byte("{\"builders\": []}\n")
	var i interface{}

	if err := Unmarshal(b, &i); err != nil {
		t.Error("Failed to unmarshal JSON")
	}
}

func TestUnmarshalYaml(t *testing.T) {
	b := []byte("---\nbuilders: []\n")
	var i interface{}

	if err := Unmarshal(b, &i); err != nil {
		t.Error("Failed to unmarshal YAML")
	}
}
