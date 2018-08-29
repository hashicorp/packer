package googlecompute

import (
	"context"

	"github.com/hashicorp/packer/helper/multistep"

	"io/ioutil"
	"os"
	"testing"
)

func TestStepCreateSSHKey_impl(t *testing.T) {
	var _ multistep.Step = new(StepCreateSSHKey)
}

func TestStepCreateSSHKey_privateKey(t *testing.T) {
	state := testState(t)
	step := new(StepCreateSSHKey)
	step.PrivateKeyFile = "test-fixtures/fake-key"
	defer step.Cleanup(state)

	// run the step
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify that we have a public/private key
	cfg := state.Get("config").(*Config)
	if len(cfg.Comm.SSHPrivateKey) == 0 {
		t.Fatal("should have key")
	}
}

func TestStepCreateSSHKey(t *testing.T) {
	state := testState(t)
	step := new(StepCreateSSHKey)
	defer step.Cleanup(state)

	// run the step
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify that we have a public/private key
	cfg := state.Get("config").(*Config)
	if len(cfg.Comm.SSHPrivateKey) == 0 {
		t.Fatal("should have key")
	}
	if len(cfg.Comm.SSHPublicKey) == 0 {
		t.Fatal("should have key")
	}
}

func TestStepCreateSSHKey_debug(t *testing.T) {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(tf.Name())
	tf.Close()

	state := testState(t)
	step := new(StepCreateSSHKey)
	step.Debug = true
	step.DebugKeyPath = tf.Name()

	defer step.Cleanup(state)

	// run the step
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify that we have a public/private key
	cfg := state.Get("config").(*Config)
	if len(cfg.Comm.SSHPrivateKey) == 0 {
		t.Fatal("should have key")
	}
	if len(cfg.Comm.SSHPublicKey) == 0 {
		t.Fatal("should have key")
	}
	if _, err := os.Stat(tf.Name()); err != nil {
		t.Fatalf("err: %s", err)
	}
}
