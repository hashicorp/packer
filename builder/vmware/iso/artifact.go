package iso

import (
	"fmt"
)

const (
	ArtifactConfFormat         = "artifact.conf.format"
	ArtifactConfKeepRegistered = "artifact.conf.keep_registered"
	ArtifactConfSkipExport     = "artifact.conf.skip_export"
)

// Artifact is the result of running the VMware builder, namely a set
// of files associated with the resulting machine.
type Artifact struct {
	builderId string
	id        string
	dir       OutputDir
	f         []string
	config    map[string]string
}

func (a *Artifact) BuilderId() string {
	return a.builderId
}

func (a *Artifact) Files() []string {
	return a.f
}

func (a *Artifact) Id() string {
	return a.id
}

func (a *Artifact) String() string {
	return fmt.Sprintf("VM files in directory: %s", a.dir)
}

func (a *Artifact) State(name string) interface{} {
	return a.config[name]
}

func (a *Artifact) Destroy() error {
	return a.dir.RemoveAll()
}
