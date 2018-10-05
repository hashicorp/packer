package schema

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestServerActionCreateImageRequest(t *testing.T) {
	var (
		oneLabel    = map[string]string{"foo": "bar"}
		nilLabels   map[string]string
		emptyLabels = map[string]string{}
	)

	testCases := []struct {
		name string
		in   ServerActionCreateImageRequest
		out  []byte
	}{
		{
			name: "no labels",
			in:   ServerActionCreateImageRequest{},
			out:  []byte(`{"type":null,"description":null}`),
		},
		{
			name: "one label",
			in:   ServerActionCreateImageRequest{Labels: &oneLabel},
			out:  []byte(`{"type":null,"description":null,"labels":{"foo":"bar"}}`),
		},
		{
			name: "nil labels",
			in:   ServerActionCreateImageRequest{Labels: &nilLabels},
			out:  []byte(`{"type":null,"description":null,"labels":null}`),
		},
		{
			name: "empty labels",
			in:   ServerActionCreateImageRequest{Labels: &emptyLabels},
			out:  []byte(`{"type":null,"description":null,"labels":{}}`),
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
