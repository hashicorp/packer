package packer

type TestHook struct {
	runCalled bool
}

func (t *TestHook) Run(string, interface{}, Ui) {
	t.runCalled = true
}
