package packer

type TestPostProcessor struct{}

func (*TestPostProcessor) Configure(interface{}) error {
	return nil
}

func (*TestPostProcessor) PostProcess(Artifact) (Artifact, error) {
	return nil, nil
}
