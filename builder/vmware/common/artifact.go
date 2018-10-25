package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer/packer"
)

// BuilderId for the local artifacts
const BuilderId = "mitchellh.vmware"
const BuilderIdESX = "mitchellh.vmware-esx"

const (
	ArtifactConfFormat         = "artifact.conf.format"
	ArtifactConfKeepRegistered = "artifact.conf.keep_registered"
	ArtifactConfSkipExport     = "artifact.conf.skip_export"
)

// Artifact is the result of running the VMware builder, namely a set
// of files associated with the resulting machine.
type artifact struct {
	builderId string
	id        string
	dir       string
	f         []string
	config    map[string]string
}

// NewLocalArtifact returns a VMware artifact containing the files
// in the given directory.
// NewLocalArtifact returns a VMware artifact containing the files
// in the given directory.
func NewLocalArtifact(id string, dir string) (packer.Artifact, error) {
	files := make([]string, 0, 5)
	visit := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	}

	if err := filepath.Walk(dir, visit); err != nil {
		return nil, err
	}

	return &artifact{
		builderId: id,
		dir:       dir,
		f:         files,
	}, nil
}

func NewArtifact(dir OutputDir, files []string, config map[string]string, esxi bool) (packer.Artifact, error) {
	builderID := BuilderId
	if esxi {
		builderID = BuilderIdESX
	}

	return &artifact{
		builderId: builderID,
		dir:       dir.String(),
		f:         files,
	}, nil
}

func (a *artifact) BuilderId() string {
	return BuilderId
}

func (a *artifact) Files() []string {
	return a.f
}

func (a *artifact) Id() string {
	return a.id
}

func (a *artifact) String() string {
	return fmt.Sprintf("VM files in directory: %s", a.dir)
}

func (a *artifact) State(name string) interface{} {
	return a.config[name]
}

func (a *artifact) Destroy() error {
	return os.RemoveAll(a.dir)
}
