package docker

import (
	"testing"
)

func TestCleanLine(t *testing.T) {
	cases := []struct {
		input  string
		output string
	}{
		{
			"\x1b[0A\x1b[2K\r8dbd9e392a96: Pulling image (precise) from ubuntu\r\x1b[0B\x1b[1A\x1b[2K\r8dbd9e392a96: Pulling image (precise) from ubuntu, endpoint: https://cdn-registry-1.docker.io/v1/\r\x1b[1B",
			"8dbd9e392a96: Pulling image (precise) from ubuntu, endpoint: https://cdn-registry-1.docker.io/v1/",
		},
	}

	for _, tc := range cases {
		actual := cleanOutputLine(tc.input)
		if actual != tc.output {
			t.Fatalf("bad: %#v %#v", tc.input, actual)
		}
	}
}
