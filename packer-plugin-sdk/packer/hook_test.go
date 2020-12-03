package packer

import (
	"context"
	"testing"
)

func TestDispatchHook_Implements(t *testing.T) {
	var _ Hook = new(DispatchHook)
}

func TestDispatchHook_Run_NoHooks(t *testing.T) {
	// Just make sure nothing blows up
	dh := &DispatchHook{}
	dh.Run(context.Background(), "foo", nil, nil, nil)
}

func TestDispatchHook_Run(t *testing.T) {
	hook := &MockHook{}

	mapping := make(map[string][]Hook)
	mapping["foo"] = []Hook{hook}
	dh := &DispatchHook{Mapping: mapping}
	dh.Run(context.Background(), "foo", nil, nil, 42)

	if !hook.RunCalled {
		t.Fatal("should be called")
	}
	if hook.RunName != "foo" {
		t.Fatalf("bad: %s", hook.RunName)
	}
	if hook.RunData != 42 {
		t.Fatalf("bad: %#v", hook.RunData)
	}
}

// A helper Hook implementation for testing cancels.
// Run will wait indetinitelly until ctx is cancelled.
type CancelHook struct {
	cancel func()
}

func (h *CancelHook) Run(ctx context.Context, _ string, _ Ui, _ Communicator, _ interface{}) error {
	h.cancel()
	<-ctx.Done()
	return ctx.Err()
}

func TestDispatchHook_cancel(t *testing.T) {

	cancelHook := new(CancelHook)

	dh := &DispatchHook{
		Mapping: map[string][]Hook{
			"foo": {cancelHook},
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancelHook.cancel = cancel

	errchan := make(chan error)
	go func() {
		errchan <- dh.Run(ctx, "foo", nil, nil, 42)
	}()

	if err := <-errchan; err == nil {
		t.Fatal("hook should've errored")
	}
}
