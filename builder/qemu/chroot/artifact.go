package chroot

import (
	"fmt"
	"os"
)

const BuilderId = "wangzw.qeum.chroot"

// Artifact is the result of running the qemu-chroot builder, namely a set
// of files associated with the resulting machine.
type Artifact struct {
	dir   string
	files []string
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return a.files
}

func (a *Artifact) Id() string {
	return a.files[0]
}

func (a *Artifact) String() string {
	return fmt.Sprintf("Image files in directory: %s", a.dir)
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	return os.RemoveAll(a.dir)
}
