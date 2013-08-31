package packer

import (
	"cgl.tideland.biz/asserts"
	"sync"
	"testing"
	"time"
)

// A helper Hook implementation for testing cancels.
type CancelHook struct {
	sync.Mutex
	cancelCh chan struct{}
	doneCh   chan struct{}

	Cancelled bool
}

func (h *CancelHook) Run(string, Ui, Communicator, interface{}) error {
	h.Lock()
	h.cancelCh = make(chan struct{})
	h.doneCh = make(chan struct{})
	h.Unlock()

	defer close(h.doneCh)

	select {
	case <-h.cancelCh:
		h.Cancelled = true
	case <-time.After(1 * time.Second):
	}

	return nil
}

func (h *CancelHook) Cancel() {
	h.Lock()
	close(h.cancelCh)
	h.Unlock()

	<-h.doneCh
}

func TestDispatchHook_Implements(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	var r Hook
	c := &DispatchHook{}

	assert.Implementor(c, &r, "should be a Hook")
}

func TestDispatchHook_Run_NoHooks(t *testing.T) {
	// Just make sure nothing blows up
	dh := &DispatchHook{}
	dh.Run("foo", nil, nil, nil)
}

func TestDispatchHook_Run(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	hook := &MockHook{}

	mapping := make(map[string][]Hook)
	mapping["foo"] = []Hook{hook}
	dh := &DispatchHook{Mapping: mapping}
	dh.Run("foo", nil, nil, 42)

	assert.True(hook.RunCalled, "run should be called")
	assert.Equal(hook.RunName, "foo", "should be proper event")
	assert.Equal(hook.RunData, 42, "should be correct data")
}

func TestDispatchHook_cancel(t *testing.T) {
	hook := new(CancelHook)

	dh := &DispatchHook{
		Mapping: map[string][]Hook{
			"foo": []Hook{hook},
		},
	}

	go dh.Run("foo", nil, nil, 42)
	time.Sleep(100 * time.Millisecond)
	dh.Cancel()

	if !hook.Cancelled {
		t.Fatal("hook should've cancelled")
	}
}
