package packer

// MockPreProcessor is an implementation of PreProcessor that can be
// used for tests.
type MockPreProcessor struct {
	Error error

	ConfigureCalled  bool
	ConfigureConfigs []interface{}
	ConfigureError   error

	PreProcessCalled bool
	PreProcessUi     Ui
}

func (t *MockPreProcessor) Configure(configs ...interface{}) error {
	t.ConfigureCalled = true
	t.ConfigureConfigs = configs
	return t.ConfigureError
}

func (t *MockPreProcessor) PreProcess(ui Ui) error {
	t.PreProcessCalled = true
	t.PreProcessUi = ui

	return t.Error
}
