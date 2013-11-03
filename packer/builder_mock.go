package packer

// MockBuilder is an implementation of Builder that can be used for tests.
// You can set some fake return values and you can keep track of what
// methods were called on the builder. It is fairly basic.
type MockBuilder struct {
	ArtifactId      string
	PrepareWarnings []string

	PrepareCalled bool
	PrepareConfig []interface{}
	RunCalled     bool
	RunCache      Cache
	RunHook       Hook
	RunUi         Ui
	CancelCalled  bool
}

func (tb *MockBuilder) Prepare(config ...interface{}) ([]string, error) {
	tb.PrepareCalled = true
	tb.PrepareConfig = config
	return tb.PrepareWarnings, nil
}

func (tb *MockBuilder) Run(ui Ui, h Hook, c Cache) (Artifact, error) {
	tb.RunCalled = true
	tb.RunHook = h
	tb.RunUi = ui
	tb.RunCache = c
	return &MockArtifact{
		IdValue: tb.ArtifactId,
	}, nil
}

func (tb *MockBuilder) Cancel() {
	tb.CancelCalled = true
}
