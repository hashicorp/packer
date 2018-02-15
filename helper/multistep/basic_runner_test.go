package multistep

import (
	"reflect"
	"testing"
	"time"
)

func TestBasicRunner_ImplRunner(t *testing.T) {
	var raw interface{}
	raw = &BasicRunner{}
	if _, ok := raw.(Runner); !ok {
		t.Fatalf("BasicRunner must be a Runner")
	}
}

func TestBasicRunner_Run(t *testing.T) {
	data := new(BasicStateBag)
	stepA := &TestStepAcc{Data: "a"}
	stepB := &TestStepAcc{Data: "b"}

	r := &BasicRunner{Steps: []Step{stepA, stepB}}
	r.Run(data)

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

	// Test no halted or cancelled
	if _, ok := data.GetOk(StateCancelled); ok {
		t.Errorf("cancelled should not be in state bag")
	}

	if _, ok := data.GetOk(StateHalted); ok {
		t.Errorf("halted should not be in state bag")
	}
}

func TestBasicRunner_Run_Halt(t *testing.T) {
	data := new(BasicStateBag)
	stepA := &TestStepAcc{Data: "a"}
	stepB := &TestStepAcc{Data: "b", Halt: true}
	stepC := &TestStepAcc{Data: "c"}

	r := &BasicRunner{Steps: []Step{stepA, stepB, stepC}}
	r.Run(data)

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

	// Test that it says it is halted
	halted := data.Get(StateHalted).(bool)
	if !halted {
		t.Errorf("not halted")
	}
}

// confirm that can't run twice
func TestBasicRunner_Run_Run(t *testing.T) {
	defer func() {
		recover()
	}()
	ch := make(chan chan bool)
	stepInt := &TestStepSync{ch}
	stepWait := &TestStepWaitForever{}
	r := &BasicRunner{Steps: []Step{stepInt, stepWait}}

	go r.Run(new(BasicStateBag))
	// wait until really running
	<-ch

	// now try to run aain
	r.Run(new(BasicStateBag))

	// should not get here in nominal codepath
	t.Errorf("Was able to run an already running BasicRunner")
}

func TestBasicRunner_Cancel(t *testing.T) {
	ch := make(chan chan bool)
	data := new(BasicStateBag)
	stepA := &TestStepAcc{Data: "a"}
	stepB := &TestStepAcc{Data: "b"}
	stepInt := &TestStepSync{ch}
	stepC := &TestStepAcc{Data: "c"}

	r := &BasicRunner{Steps: []Step{stepA, stepB, stepInt, stepC}}

	// cancelling an idle Runner is a no-op
	r.Cancel()

	go r.Run(data)

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

func TestBasicRunner_Cancel_Special(t *testing.T) {
	stepOne := &TestStepInjectCancel{}
	stepTwo := &TestStepInjectCancel{}
	r := &BasicRunner{Steps: []Step{stepOne, stepTwo}}

	state := new(BasicStateBag)
	state.Put("runner", r)
	r.Run(state)

	// test that state contains cancelled
	if _, ok := state.GetOk(StateCancelled); !ok {
		t.Errorf("cancelled should be in state bag")
	}
}
