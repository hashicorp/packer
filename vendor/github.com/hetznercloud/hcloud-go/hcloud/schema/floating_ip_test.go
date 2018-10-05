package schema

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestFloatingIPCreateRequest(t *testing.T) {
	var (
		oneLabel    = map[string]string{"foo": "bar"}
		nilLabels   map[string]string
		emptyLabels = map[string]string{}
	)

	testCases := []struct {
		name string
		in   FloatingIPCreateRequest
		out  []byte
	}{
		{
			name: "no labels",
			in:   FloatingIPCreateRequest{Type: "ipv4"},
			out:  []byte(`{"type":"ipv4"}`),
		},
		{
			name: "one label",
			in:   FloatingIPCreateRequest{Type: "ipv4", Labels: &oneLabel},
			out:  []byte(`{"type":"ipv4","labels":{"foo":"bar"}}`),
		},
		{
			name: "nil labels",
			in:   FloatingIPCreateRequest{Type: "ipv4", Labels: &nilLabels},
			out:  []byte(`{"type":"ipv4","labels":null}`),
		},
		{
			name: "empty labels",
			in:   FloatingIPCreateRequest{Type: "ipv4", Labels: &emptyLabels},
			out:  []byte(`{"type":"ipv4","labels":{}}`),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			data, err := json.Marshal(testCase.in)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(data, testCase.out) {
				t.Fatalf("output %s does not match %s", data, testCase.out)
			}
		})
	}
}
