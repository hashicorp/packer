// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package manifest

import "fmt"

const BuilderId = "packer.post-processor.manifest"

type ArtifactFile struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type Artifact struct {
	BuildName     string            `json:"name"`
	BuilderType   string            `json:"builder_type"`
	BuildTime     int64             `json:"build_time,omitempty"`
	ArtifactFiles []ArtifactFile    `json:"files"`
	ArtifactId    string            `json:"artifact_id"`
	PackerRunUUID string            `json:"packer_run_uuid"`
	CustomData    map[string]string `json:"custom_data"`
}

func (a *Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	var files []string
	for _, af := range a.ArtifactFiles {
		files = append(files, af.Name)
	}
	return files
}

func (a *Artifact) Id() string {
	return a.ArtifactId
}

func (a *Artifact) String() string {
	return fmt.Sprintf("%s-%s", a.BuildName, a.ArtifactId)
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	return nil
}
