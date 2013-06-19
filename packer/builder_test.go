package packer

type TestBuilder struct {
	artifactId string

	prepareCalled bool
	prepareConfig []interface{}
	runCalled     bool
	runCache      Cache
	runHook       Hook
	runUi         Ui
	cancelCalled  bool
}

func (tb *TestBuilder) Prepare(config ...interface{}) error {
	tb.prepareCalled = true
	tb.prepareConfig = config
	return nil
}

func (tb *TestBuilder) Run(ui Ui, h Hook, c Cache) (Artifact, error) {
	tb.runCalled = true
	tb.runHook = h
	tb.runUi = ui
	tb.runCache = c
	return &TestArtifact{id: tb.artifactId}, nil
}

func (tb *TestBuilder) Cancel() {
	tb.cancelCalled = true
}
