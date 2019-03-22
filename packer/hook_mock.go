package packer

import (
	"context"
	"time"
)

// MockHook is an implementation of Hook that can be used for tests.
type MockHook struct {
	RunFunc func() error

	RunCalled    bool
	RunComm      Communicator
	RunData      interface{}
	RunName      string
	RunUi        Ui
	CancelCalled bool
}

func (t *MockHook) Run(ctx context.Context, name string, ui Ui, comm Communicator, data interface{}) error {

	go func() {
		select {
		case <-time.After(2 * time.Minute):
		case <-ctx.Done():
			t.CancelCalled = true
		}
	}()

	t.RunCalled = true
	t.RunComm = comm
	t.RunData = data
	t.RunName = name
	t.RunUi = ui

	if t.RunFunc == nil {
		return nil
	}

	return t.RunFunc()
}
