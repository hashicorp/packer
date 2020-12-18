package packer

import (
	"context"
)

// MockHook is an implementation of Hook that can be used for tests.
type MockHook struct {
	RunFunc func(context.Context) error

	RunCalled bool
	RunComm   Communicator
	RunData   interface{}
	RunName   string
	RunUi     Ui
}

func (t *MockHook) Run(ctx context.Context, name string, ui Ui, comm Communicator, data interface{}) error {

	t.RunCalled = true
	t.RunComm = comm
	t.RunData = data
	t.RunName = name
	t.RunUi = ui

	if t.RunFunc == nil {
		return nil
	}

	return t.RunFunc(ctx)
}
