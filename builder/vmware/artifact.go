package vmware

import (
	"fmt"
	"os"
)

// Artifact is the result of running the VMware builder, namely a set
// of files associated with the resulting machine.
type Artifact struct {
	builderId string
	dir       string
	f         []string
}

func (a *Artifact) BuilderId() string {
	return a.builderId
}

func (a *Artifact) Files() []string {
	return a.f
}

func (*Artifact) Id() string {
	return "VM"
}

func (a *Artifact) String() string {
	return fmt.Sprintf("VM files in directory: %s", a.dir)
}

func (a *Artifact) Destroy() error {
	return os.RemoveAll(a.dir)
}
