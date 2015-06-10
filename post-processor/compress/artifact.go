package compress

import (
	"fmt"
	"os"
)

const BuilderId = "packer.post-processor.compress"

type Artifact struct {
	builderId string
	dir       string
	f         []string
}

func (a *Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return a.f
}

func (*Artifact) Id() string {
	return "COMPRESS"
}

func (a *Artifact) String() string {
	return fmt.Sprintf("VM compressed files in directory: %s", a.dir)
}

func (*Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	return os.RemoveAll(a.dir)
}
