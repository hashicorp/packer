package oci

import (
	"errors"
	"fmt"
	"testing"
)

const (
	ValidID = "ID"
)

type testWaitSvc struct {
	states []string
	idx    int
	err    error
}

func (tw *testWaitSvc) GetResourceState(id string) (string, error) {
	if id != ValidID {
		return "", fmt.Errorf("Invalid id %s", id)
	}
	if tw.err != nil {
		return "", tw.err
	}

	if tw.idx >= len(tw.states) {
		panic("Invalid test state")
	}
	state := tw.states[tw.idx]
	tw.idx++
	return state, nil
}

func TestReturnsWhenWaitStateIsReachedImmediately(t *testing.T) {
	ws := &testWaitSvc{states: []string{"OK"}}
	w := NewWaiter()
	err := w.WaitForResourceToReachState(ws, ValidID, []string{}, "OK")
	if err != nil {
		t.Errorf("Failed to reach expected state, got %s", err)
	}
}

func TestReturnsWhenResourceWaitsInValidWaitingState(t *testing.T) {
	w := &Waiter{WaitDurationMS: 1, MaxRetries: defaultMaxRetries}
	ws := &testWaitSvc{states: []string{"WAITING", "OK"}}
	err := w.WaitForResourceToReachState(ws, ValidID, []string{"WAITING"}, "OK")
	if err != nil {
		t.Errorf("Failed to reach expected state, got %s", err)
	}
}

func TestPropagatesErrorFromGetter(t *testing.T) {
	w := NewWaiter()
	ws := &testWaitSvc{states: []string{}, err: errors.New("ERROR")}
	err := w.WaitForResourceToReachState(ws, ValidID, []string{"WAITING"}, "OK")
	if err != ws.err {
		t.Errorf("Expected error from getter got %s", err)
	}
}

func TestReportsInvalidTransitionStateAsError(t *testing.T) {
	w := NewWaiter()
	tw := &testWaitSvc{states: []string{"UNKNOWN_STATE"}, err: errors.New("ERROR")}
	err := w.WaitForResourceToReachState(tw, ValidID, []string{"WAITING"}, "OK")
	if err == nil {
		t.Fatal("Expected error from getter")
	}
}

func TestErrorsWhenMaxWaitTriesExceeded(t *testing.T) {
	w := Waiter{WaitDurationMS: 1, MaxRetries: 1}

	ws := &testWaitSvc{states: []string{"WAITING", "OK"}}

	err := w.WaitForResourceToReachState(ws, ValidID, []string{"WAITING"}, "OK")
	if err == nil {
		t.Fatal("Expecting error but wait terminated")
	}
}
