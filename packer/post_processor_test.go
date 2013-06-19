package packer

type TestPostProcessor struct {
	configCalled bool
	configVal    interface{}
	ppCalled     bool
	ppArtifact   Artifact
	ppUi         Ui
}

func (pp *TestPostProcessor) Configure(v interface{}) error {
	pp.configCalled = true
	pp.configVal = v
	return nil
}

func (pp *TestPostProcessor) PostProcess(ui Ui, a Artifact) (Artifact, error) {
	pp.ppCalled = true
	pp.ppArtifact = a
	pp.ppUi = ui
	return new(TestArtifact), nil
}
