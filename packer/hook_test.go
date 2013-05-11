package packer

type TestHook struct {
	runCalled bool
}

func (t *TestHook) Run(string, interface{}) {
	t.runCalled = true
}
