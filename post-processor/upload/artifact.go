package upload

import "fmt"

const BuilderId = "packer.post-processor.upload"

type Artifact struct {
	endpoint string
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (*Artifact) Files() []string {
	return []string{}
}

func (*Artifact) Id() string {
	return ""
}

func (*Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) String() string {
	return fmt.Sprintf("files available at: %s", a.endpoint)
}

func (a *Artifact) Destroy() error {
	return nil
}
