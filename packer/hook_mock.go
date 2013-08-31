package packer

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

func (t *MockHook) Run(name string, ui Ui, comm Communicator, data interface{}) error {
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

func (t *MockHook) Cancel() {
	t.CancelCalled = true
}
