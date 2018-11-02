package common

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

const (
	// BuilderId for the local artifacts
	BuilderId    = "mitchellh.vmware"
	BuilderIdESX = "mitchellh.vmware-esx"

	ArtifactConfFormat         = "artifact.conf.format"
	ArtifactConfKeepRegistered = "artifact.conf.keep_registered"
	ArtifactConfSkipExport     = "artifact.conf.skip_export"
)

// Artifact is the result of running the VMware builder, namely a set
// of files associated with the resulting machine.
type artifact struct {
	builderId string
	id        string
	dir       OutputDir
	f         []string
	config    map[string]string
}

func (a *artifact) BuilderId() string {
	return a.builderId
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
	return a.dir.RemoveAll()
}

func NewArtifact(remoteType string, format string, exportOutputPath string, vmName string, skipExport bool, keepRegistered bool, state multistep.StateBag) (packer.Artifact, error) {
	var files []string
	var dir OutputDir
	var err error
	if remoteType != "" && !skipExport {
		dir = new(LocalOutputDir)
		dir.SetOutputDir(exportOutputPath)
		files, err = dir.ListFiles()
	} else {
		files, err = state.Get("dir").(OutputDir).ListFiles()
	}
	if err != nil {
		return nil, err
	}

	// Set the proper builder ID
	builderId := BuilderId
	if remoteType != "" {
		builderId = BuilderIdESX
	}

	config := make(map[string]string)
	config[ArtifactConfKeepRegistered] = strconv.FormatBool(keepRegistered)
	config[ArtifactConfFormat] = format
	config[ArtifactConfSkipExport] = strconv.FormatBool(skipExport)

	return &artifact{
		builderId: builderId,
		id:        vmName,
		dir:       dir,
		f:         files,
		config:    config,
	}, nil
}
