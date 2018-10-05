package schema

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestSSHKeyCreateRequest(t *testing.T) {
	var (
		oneLabel    = map[string]string{"foo": "bar"}
		nilLabels   map[string]string
		emptyLabels = map[string]string{}
	)

	testCases := []struct {
		name string
		in   SSHKeyCreateRequest
		out  []byte
	}{
		{
			name: "no labels",
			in:   SSHKeyCreateRequest{Name: "test", PublicKey: "key"},
			out:  []byte(`{"name":"test","public_key":"key"}`),
		},
		{
			name: "one label",
			in:   SSHKeyCreateRequest{Name: "test", PublicKey: "key", Labels: &oneLabel},
			out:  []byte(`{"name":"test","public_key":"key","labels":{"foo":"bar"}}`),
		},
		{
			name: "nil labels",
			in:   SSHKeyCreateRequest{Name: "test", PublicKey: "key", Labels: &nilLabels},
			out:  []byte(`{"name":"test","public_key":"key","labels":null}`),
		},
		{
			name: "empty labels",
			in:   SSHKeyCreateRequest{Name: "test", PublicKey: "key", Labels: &emptyLabels},
			out:  []byte(`{"name":"test","public_key":"key","labels":{}}`),
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
