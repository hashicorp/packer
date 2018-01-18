package multistep

import (
	"os"
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestDebugRunner_Impl(t *testing.T) {
	var raw interface{}
	raw = &DebugRunner{}
	if _, ok := raw.(Runner); !ok {
		t.Fatal("DebugRunner must be a runner.")
	}
}

func TestDebugRunner_Run(t *testing.T) {
	data := new(BasicStateBag)
	stepA := &TestStepAcc{Data: "a"}
	stepB := &TestStepAcc{Data: "b"}

	pauseFn := func(loc DebugLocation, name string, state StateBag) {
		key := "data"
		if loc == DebugLocationBeforeCleanup {
			key = "cleanup"
		}

		if _, ok := state.GetOk(key); !ok {
			state.Put(key, make([]string, 0, 5))
		}

		data := state.Get(key).([]string)
		state.Put(key, append(data, name))
	}

	r := &DebugRunner{
		Steps:   []Step{stepA, stepB},
		PauseFn: pauseFn,
	}

	r.Run(context.Background(), data)

	// Test data
	expected := []string{"a", "TestStepAcc", "b", "TestStepAcc"}
	results := data.Get("data").([]string)
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("unexpected results: %#v", results)
	}

	// Test cleanup
	expected = []string{"TestStepAcc", "b", "TestStepAcc", "a"}
	results = data.Get("cleanup").([]string)
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("unexpected results: %#v", results)
	}
}

// confirm that can't run twice
func TestDebugRunner_Run_Run(t *testing.T) {
	defer func() {
		recover()
	}()
	ch := make(chan chan bool)
	stepInt := &TestStepSync{ch}
	stepWait := &TestStepWaitForever{}
	r := &DebugRunner{Steps: []Step{stepInt, stepWait}}

	go r.Run(context.Background(), new(BasicStateBag))
	// wait until really running
	<-ch

	// now try to run aain
	r.Run(context.Background(), new(BasicStateBag))

	// should not get here in nominal codepath
	t.Errorf("Was able to run an already running DebugRunner")
}

func TestDebugRunner_Cancel(t *testing.T) {
	ch := make(chan chan bool)
	data := new(BasicStateBag)
	stepA := &TestStepAcc{Data: "a"}
	stepB := &TestStepAcc{Data: "b"}
	stepInt := &TestStepSync{ch}
	stepC := &TestStepAcc{Data: "c"}

	r := &DebugRunner{}
	r.Steps = []Step{stepA, stepB, stepInt, stepC}

	// cancelling an idle Runner is a no-op
	r.Cancel()

	go r.Run(context.Background(), data)

	// Wait until we reach the sync point
	responseCh := <-ch

	// Cancel then continue chain
	cancelCh := make(chan bool)
	go func() {
		r.Cancel()
		cancelCh <- true
	}()

	for {
		if _, ok := data.GetOk(StateCancelled); ok {
			responseCh <- true
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	<-cancelCh

	// Test run data
	expected := []string{"a", "b"}
	results := data.Get("data").([]string)
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("unexpected result: %#v", results)
	}

	// Test cleanup data
	expected = []string{"b", "a"}
	results = data.Get("cleanup").([]string)
	if !reflect.DeepEqual(results, expected) {
		t.Errorf("unexpected result: %#v", results)
	}

	// Test that it says it is cancelled
	cancelled := data.Get(StateCancelled).(bool)
	if !cancelled {
		t.Errorf("not cancelled")
	}
}

func TestDebugPauseDefault(t *testing.T) {

	// Create a pipe pair so that writes/reads are blocked until we do it
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Set stdin so we can control it
	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	// Start pausing
	complete := make(chan bool, 1)
	go func() {
		dr := &DebugRunner{Steps: []Step{
			&TestStepAcc{Data: "a"},
		}}
		dr.Run(context.Background(), new(BasicStateBag))
		complete <- true
	}()

	select {
	case <-complete:
		t.Fatal("shouldn't have completed")
	case <-time.After(100 * time.Millisecond):
	}

	w.Write([]byte("\n\n"))

	select {
	case <-complete:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("didn't complete")
	}
}
