package packer

type TestBuilder struct {
	prepareCalled bool
	prepareConfig interface{}
	runCalled     bool
	runCache      Cache
	runHook       Hook
	runUi         Ui
	cancelCalled  bool
}

func (tb *TestBuilder) Prepare(config interface{}) error {
	tb.prepareCalled = true
	tb.prepareConfig = config
	return nil
}

func (tb *TestBuilder) Run(ui Ui, h Hook, c Cache) Artifact {
	tb.runCalled = true
	tb.runHook = h
	tb.runUi = ui
	tb.runCache = c
	return nil
}

func (tb *TestBuilder) Cancel() {
	tb.cancelCalled = true
}
