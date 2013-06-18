package packer

type TestArtifact struct{}

func (*TestArtifact) BuilderId() string {
	return "bid"
}

func (*TestArtifact) Files() []string {
	return []string{"a", "b"}
}

func (*TestArtifact) Id() string {
	return "id"
}

func (*TestArtifact) String() string {
	return "string"
}

func (*TestArtifact) Destroy() error {
	return nil
}
