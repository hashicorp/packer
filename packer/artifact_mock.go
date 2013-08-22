package packer

// MockArtifact is an implementation of Artifact that can be used for tests.
type MockArtifact struct {
	IdValue       string
	DestroyCalled bool
}

func (*MockArtifact) BuilderId() string {
	return "bid"
}

func (*MockArtifact) Files() []string {
	return []string{"a", "b"}
}

func (a *MockArtifact) Id() string {
	id := a.IdValue
	if id == "" {
		id = "id"
	}

	return id
}

func (*MockArtifact) String() string {
	return "string"
}

func (a *MockArtifact) Destroy() error {
	a.DestroyCalled = true
	return nil
}
