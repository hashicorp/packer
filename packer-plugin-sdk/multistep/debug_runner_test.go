package multistep

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"
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

type TestStepFn struct {
	run     func(context.Context, StateBag) StepAction
	cleanup func(StateBag)
}

var _ Step = TestStepFn{}

func (fn TestStepFn) Run(ctx context.Context, sb StateBag) StepAction {
	return fn.run(ctx, sb)
}

func (fn TestStepFn) Cleanup(sb StateBag) {
	if fn.cleanup != nil {
		fn.cleanup(sb)
	}
}
func TestDebugRunner_Cancel(t *testing.T) {

	topCtx, topCtxCancel := context.WithCancel(context.Background())

	checkCancelled := func(data StateBag) {
		cancelled := data.Get(StateCancelled).(bool)
		if !cancelled {
			t.Fatal("state should be cancelled")
		}
	}

	data := new(BasicStateBag)
	r := &DebugRunner{}
	r.Steps = []Step{
		&TestStepAcc{Data: "a"},
		&TestStepAcc{Data: "b"},
		TestStepFn{
			run: func(ctx context.Context, sb StateBag) StepAction {
				return ActionContinue
			},
			cleanup: checkCancelled,
		},
		TestStepFn{
			run: func(ctx context.Context, sb StateBag) StepAction {
				topCtxCancel()
				<-ctx.Done()
				return ActionContinue
			},
			cleanup: checkCancelled,
		},
		TestStepFn{
			run: func(context.Context, StateBag) StepAction {
				t.Fatal("I should not be called")
				return ActionContinue
			},
			cleanup: func(StateBag) {
				t.Fatal("I should not be called")
			},
		},
	}

	r.Run(topCtx, data)

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
	cancelled, ok := data.GetOk(StateCancelled)
	if !ok {
		t.Fatal("could not get state cancelled")
	}
	if !cancelled.(bool) {
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
