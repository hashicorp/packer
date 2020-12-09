package common

import (
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
)

func TestCommHost(t *testing.T) {
	state := testState(t)
	config := SSHConfig{
		Comm: communicator.Config{
			Type: "ssh",
			SSH: communicator.SSH{
				SSHHost: "127.0.0.1",
			},
		},
	}

	hostFunc := CommHost(&config)
	out, err := hostFunc(state)
	if err != nil {
		t.Fatalf("Should not have had an error")
	}

	if out != "127.0.0.1" {
		t.Fatalf("Should have respected ssh override.")
	}
}
