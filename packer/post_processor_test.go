package packer

type TestPostProcessor struct {
	configCalled bool
	configVal    interface{}
	ppCalled     bool
	ppArtifact   Artifact
}

func (pp *TestPostProcessor) Configure(v interface{}) error {
	pp.configCalled = true
	pp.configVal = v
	return nil
}

func (pp *TestPostProcessor) PostProcess(a Artifact) (Artifact, error) {
	pp.ppCalled = true
	pp.ppArtifact = a
	return nil, nil
}
