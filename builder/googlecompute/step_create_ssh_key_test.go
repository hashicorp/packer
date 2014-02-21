package googlecompute

import (
	"github.com/mitchellh/multistep"

	"io/ioutil"
	"os"
	"testing"
)

func TestStepCreateSSHKey_impl(t *testing.T) {
	var _ multistep.Step = new(StepCreateSSHKey)
}

func TestStepCreateSSHKey(t *testing.T) {
	state := testState(t)
	step := new(StepCreateSSHKey)
	defer step.Cleanup(state)

	// run the step
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify that we have a public/private key
	if _, ok := state.GetOk("ssh_private_key"); !ok {
		t.Fatal("should have key")
	}
	if _, ok := state.GetOk("ssh_public_key"); !ok {
		t.Fatal("should have key")
	}
}

func TestStepCreateSSHKey_debug(t *testing.T) {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Close()

	state := testState(t)
	step := new(StepCreateSSHKey)
	step.Debug = true
	step.DebugKeyPath = tf.Name()

	defer step.Cleanup(state)

	// run the step
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify that we have a public/private key
	if _, ok := state.GetOk("ssh_private_key"); !ok {
		t.Fatal("should have key")
	}
	if _, ok := state.GetOk("ssh_public_key"); !ok {
		t.Fatal("should have key")
	}
	if _, err := os.Stat(tf.Name()); err != nil {
		t.Fatalf("err: %s", err)
	}
}
