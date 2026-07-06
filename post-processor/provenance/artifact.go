// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package provenance

import "fmt"

const BuilderId = "packer.post-processor.provenance"

type Artifact struct {
	ArtifactID string   `json:"artifact_id"`
	FilesList  []string `json:"files"`
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return append([]string(nil), a.FilesList...)
}

func (a *Artifact) Id() string {
	return a.ArtifactID
}

func (a *Artifact) String() string {
	return fmt.Sprintf("provenance sidecars: %v", a.FilesList)
}

func (*Artifact) State(string) interface{} {
	return nil
}

func (*Artifact) Destroy() error {
	return nil
}
