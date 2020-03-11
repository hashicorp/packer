package packer

// MockArtifact is an implementation of Artifact that can be used for tests.
type MockArtifact struct {
	BuilderIdValue string
	FilesValue     []string
	IdValue        string
	StateValues    map[string]interface{}
	DestroyCalled  bool
	StringValue    string
}

func (a *MockArtifact) BuilderId() string {
	if a.BuilderIdValue == "" {
		return "bid"
	}

	return a.BuilderIdValue
}

func (a *MockArtifact) Files() []string {
	if a.FilesValue == nil {
		return []string{"a", "b"}
	}

	return a.FilesValue
}

func (a *MockArtifact) Id() string {
	id := a.IdValue
	if id == "" {
		id = "id"
	}

	return id
}

func (a *MockArtifact) String() string {
	str := a.StringValue
	if str == "" {
		str = "string"
	}
	return str
}

func (a *MockArtifact) State(name string) interface{} {
	value := a.StateValues[name]
	return value
}

func (a *MockArtifact) Destroy() error {
	a.DestroyCalled = true
	return nil
}
