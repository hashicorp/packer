package packer

// MockPostProcessor is an implementation of PostProcessor that can be
// used for tests.
type MockPostProcessor struct {
	ArtifactId string
	Keep       bool
	Error      error

	ConfigureCalled  bool
	ConfigureConfigs []interface{}
	ConfigureError   error

	PostProcessCalled   bool
	PostProcessArtifact Artifact
	PostProcessUi       Ui
}

func (t *MockPostProcessor) Configure(configs ...interface{}) error {
	t.ConfigureCalled = true
	t.ConfigureConfigs = configs
	return t.ConfigureError
}

func (t *MockPostProcessor) PostProcess(ui Ui, a Artifact) (Artifact, bool, error) {
	t.PostProcessCalled = true
	t.PostProcessArtifact = a
	t.PostProcessUi = ui

	return &MockArtifact{
		IdValue: t.ArtifactId,
	}, t.Keep, t.Error
}
