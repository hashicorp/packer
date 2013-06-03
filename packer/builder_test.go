package packer

type TestBuilder struct {
	prepareCalled bool
	prepareConfig interface{}
	runCalled     bool
	runHook       Hook
	runUi         Ui
	cancelCalled  bool
}

func (tb *TestBuilder) Prepare(config interface{}) error {
	tb.prepareCalled = true
	tb.prepareConfig = config
	return nil
}

func (tb *TestBuilder) Run(ui Ui, h Hook) Artifact {
	tb.runCalled = true
	tb.runHook = h
	tb.runUi = ui
	return nil
}

func (tb *TestBuilder) Cancel() {
	tb.cancelCalled = true
}
