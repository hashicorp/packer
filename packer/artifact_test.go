package packer

type TestArtifact struct {
	id            string
	destroyCalled bool
}

func (*TestArtifact) BuilderId() string {
	return "bid"
}

func (*TestArtifact) Files() []string {
	return []string{"a", "b"}
}

func (a *TestArtifact) Id() string {
	id := a.id
	if id == "" {
		id = "id"
	}

	return id
}

func (*TestArtifact) String() string {
	return "string"
}

func (a *TestArtifact) Destroy() error {
	a.destroyCalled = true
	return nil
}
