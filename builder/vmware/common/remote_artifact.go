package common

import (
	"fmt"
)

// Artifact is the result of running the VMware builder, namely a set
// of files associated with the resulting machine.
type RemoteArtifact struct {
	builderId string
	id        string
	dir       OutputDir
	f         []string
	config    map[string]string
}

func (a *RemoteArtifact) BuilderId() string {
	return a.builderId
}

func (a *RemoteArtifact) Files() []string {
	return a.f
}

func (a *RemoteArtifact) Id() string {
	return a.id
}

func (a *RemoteArtifact) String() string {
	return fmt.Sprintf("VM files in directory: %s", a.dir)
}

func (a *RemoteArtifact) State(name string) interface{} {
	return a.config[name]
}

func (a *RemoteArtifact) Destroy() error {
	return a.dir.RemoveAll()
}
