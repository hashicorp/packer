package exoscaleimport

const BuilderId = "packer.post-processor.exoscale-import"

type Artifact struct {
	id string
}

func (a *Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Id() string {
	return a.id
}

func (a *Artifact) Files() []string {
	return nil
}

func (a *Artifact) String() string {
	return a.id
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	return nil
}
