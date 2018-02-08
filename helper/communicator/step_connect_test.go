package communicator

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/stretchr/testify/assert"
)

func TestStepConnect_impl(t *testing.T) {
	var _ multistep.Step = new(StepConnect)
}

func TestStepConnect_none(t *testing.T) {
	state := testState(t)

	step := &StepConnect{
		Config: &Config{
			Type: "none",
		},
	}
	defer step.Cleanup(state)

	// run the step
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
}

var noProxyTests = []struct {
	current  string
	expected string
}{
	{"", "foo:1"},
	{"foo:1", "foo:1"},
	{"foo:1,bar:2", "foo:1,bar:2"},
	{"bar:2", "bar:2,foo:1"},
}

func TestStepConnect_setNoProxy(t *testing.T) {
	key := "NO_PROXY"
	for _, tt := range noProxyTests {
		if value := os.Getenv(key); value != "" {
			os.Unsetenv(key)
			defer func() { os.Setenv(key, value) }()
			os.Setenv(key, tt.current)
			assert.Equal(t, tt.expected, os.Getenv(key), "env not set correctly.")
		}
	}
}

func testState(t *testing.T) multistep.StateBag {
	state := new(multistep.BasicStateBag)
	state.Put("hook", &packer.MockHook{})
	state.Put("ui", &packer.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	return state
}
